package grpc

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/apus-run/gaea/internal/matcher"
	pb "github.com/apus-run/gaea/internal/testdata/helloworld"
	"github.com/apus-run/gaea/middleware"
)

// service is used to implement helloworld.GreeterServer.
type service struct {
	pb.UnimplementedGreeterServer
}

func (s *service) SayHelloStream(streamServer pb.Greeter_SayHelloStreamServer) error {
	var cnt uint
	for {
		in, err := streamServer.Recv()
		if err != nil {
			return err
		}
		if in.Name == "error" {
			panic(fmt.Sprintf("invalid argument %s", in.Name))
		}
		if in.Name == "panic" {
			panic("server panic")
		}
		err = streamServer.Send(&pb.HelloReply{
			Message: fmt.Sprintf("hello %s", in.Name),
		})
		if err != nil {
			return err
		}
		cnt++
		if cnt > 1 {
			return nil
		}
	}
}

// SayHello implements helloworld.GreeterServer
func (s *service) SayHello(_ context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	if in.Name == "error" {
		panic(fmt.Sprintf("invalid argument %s", in.Name))
	}
	if in.Name == "panic" {
		panic("server panic")
	}
	return &pb.HelloReply{Message: fmt.Sprintf("Hello %+v", in.Name)}, nil
}

type testKey struct{}

func TestServer(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, testKey{}, "test")
	srv := NewServer(
		Middleware(
			func(handler middleware.Handler) middleware.Handler {
				return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
					return handler(ctx, req)
				}
			}),
		UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			return handler(ctx, req)
		}),
		GrpcOptions(grpc.InitialConnWindowSize(0)),
	)
	pb.RegisterGreeterServer(srv, &service{})

	if e, err := srv.Endpoint(); err != nil || e == nil || strings.HasSuffix(e.Host, ":0") {
		t.Fatal(e, err)
	}

	go func() {
		// start server
		if err := srv.Start(ctx); err != nil {
			panic(err)
		}
	}()
	time.Sleep(time.Second)
	testClient(t, srv)
	_ = srv.Stop(ctx)
}

func testClient(t *testing.T, srv *Server) {
	u, err := srv.Endpoint()
	if err != nil {
		t.Fatal(err)
	}
	// new a gRPC client
	conn, err := DialInsecure(context.Background(), WithEndpoint(u.Host),
		WithDialOptions(grpc.WithBlock()),
		WithUnaryInterceptor(
			func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				return invoker(ctx, method, req, reply, cc, opts...)
			}),
		WithMiddleware(func(handler middleware.Handler) middleware.Handler {
			return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
				return handler(ctx, req)
			}
		}),
	)
	defer func() {
		_ = conn.Close()
	}()
	if err != nil {
		t.Fatal(err)
	}
	client := pb.NewGreeterClient(conn)
	reply, err := client.SayHello(context.Background(), &pb.HelloRequest{Name: "gaea"})
	t.Log(err)
	if err != nil {
		t.Errorf("failed to call: %v", err)
	}
	if !reflect.DeepEqual(reply.Message, "Hello gaea") {
		t.Errorf("expect %s, got %s", "Hello gaea", reply.Message)
	}

	streamCli, err := client.SayHelloStream(context.Background())
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		_ = streamCli.CloseSend()
	}()
	err = streamCli.Send(&pb.HelloRequest{Name: "cc"})
	if err != nil {
		t.Error(err)
		return
	}
	reply, err = streamCli.Recv()
	if err != nil {
		t.Error(err)
		return
	}
	if !reflect.DeepEqual(reply.Message, "hello cc") {
		t.Errorf("expect %s, got %s", "hello cc", reply.Message)
	}
}

type testResp struct {
	Data string
}

func TestServer_unaryServerInterceptor(t *testing.T) {
	u, err := url.Parse("grpc://hello/world")
	if err != nil {
		t.Errorf("expect %v, got %v", nil, err)
	}
	srv := &Server{
		ctx:        context.Background(),
		endpoint:   u,
		timeout:    time.Duration(10),
		middleware: matcher.New(),
	}
	srv.middleware.Use(EmptyMiddleware())
	req := &struct{}{}
	rv, err := srv.unaryServerInterceptor()(context.TODO(), req, &grpc.UnaryServerInfo{},
		func(ctx context.Context, req interface{}) (i interface{}, e error) {
			return &testResp{Data: "hi"}, nil
		})
	if err != nil {
		t.Errorf("expect %v, got %v", nil, err)
	}
	if !reflect.DeepEqual("hi", rv.(*testResp).Data) {
		t.Errorf("expect %s, got %s", "hi", rv.(*testResp).Data)
	}
}
