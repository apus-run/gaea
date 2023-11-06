package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"

	serverGrpc "github.com/apus-run/gaea/server/grpc"
	gwRuntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	log "google.golang.org/grpc/grpclog"

	helloworldpb "github.com/apus-run/gaea/internal/testdata/helloworld"
	"github.com/apus-run/gaea/middleware/recovery"
)

type server struct {
	helloworldpb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *helloworldpb.HelloRequest) (*helloworldpb.HelloReply, error) {
	if in.Name == "" {
		return nil, fmt.Errorf("invalid argument %s", in.Name)
	}
	return &helloworldpb.HelloReply{Message: fmt.Sprintf("Hello %+v", in.Name)}, nil
}

func runServer(stop <-chan struct{}) {
	go func() {
		lr, err := net.Listen("tcp", ":9998")
		if err != nil {
			panic(err)
		}

		s := serverGrpc.NewServer()

		helloworldpb.RegisterGreeterServer(s, &server{})

		if err := s.Serve(lr); err != nil {
			panic(err)
		}
	}()

	go func() {
		mux := http.NewServeMux()

		//gwHandler, err := CreateGatewayEndpoint(
		//	context.TODO(),
		//	":9998",
		//	make([]gwRuntime.ServeMuxOption, 0),
		//	[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
		//	[]AnnotatorFunc{ParalusGatewayAnnotator},
		//	helloworldpb.RegisterGreeterHandlerFromEndpoint,
		//)

		endpoint := ":9998"
		conn, err := serverGrpc.DialInsecure(
			context.Background(),
			serverGrpc.WithEndpoint(endpoint),
			serverGrpc.WithMiddleware(
				recovery.Recovery(),
			),
		)

		defer conn.Close()

		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}

		paralusJSON := NewParalusJSON()
		gwHandler, err := CreateGateway(
			context.TODO(),
			conn,
			[]gwRuntime.ServeMuxOption{
				gwRuntime.WithMarshalerOption(jsonContentType, paralusJSON),
			},
			nil,
			helloworldpb.RegisterGreeterHandler,
		)

		mux.Handle("/", gwHandler)

		hs := http.Server{
			Addr:    ":9999",
			Handler: mux,
		}

		if err = hs.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	<-stop
}

func runGateway(stop <-chan struct{}) {
	go func() {
		lr, err := net.Listen("tcp", ":9998")
		if err != nil {
			panic(err)
		}

		s := serverGrpc.NewServer()

		helloworldpb.RegisterGreeterServer(s, &server{})

		if err := s.Serve(lr); err != nil {
			panic(err)
		}
	}()
	go func() {
		ctx := context.Background()
		endpoint := ":9998"
		conn, err := serverGrpc.DialInsecure(
			ctx,
			serverGrpc.WithEndpoint(endpoint),
			serverGrpc.WithMiddleware(
				recovery.Recovery(),
			),
		)
		defer func() {
			if err != nil {
				if cerr := conn.Close(); cerr != nil {
					log.Infof("Failed to close conn to %s: %v", endpoint, cerr)
				}
				return
			}
			go func() {
				<-ctx.Done()
				if cerr := conn.Close(); cerr != nil {
					log.Infof("Failed to close conn to %s: %v", endpoint, cerr)
				}
			}()
		}()

		paralusJSON := NewParalusJSON()
		err = Run(
			ctx,
			WithAddress(":9999"),
			WithConn(conn),
			WithServeMuxOptions(gwRuntime.WithMarshalerOption(jsonContentType, paralusJSON)),
			WithAnnotator(ParalusGatewayAnnotator),
			WithHandlers(helloworldpb.RegisterGreeterHandler),
		)
		if err != nil {
			log.Fatal(err)
		}
	}()
	<-stop
}

func runGatewayV1(stop <-chan struct{}) {
	go func() {
		lr, err := net.Listen("tcp", ":9998")
		if err != nil {
			panic(err)
		}

		s := serverGrpc.NewServer()

		helloworldpb.RegisterGreeterServer(s, &server{})

		if err := s.Serve(lr); err != nil {
			panic(err)
		}
	}()
	go func() {
		ctx := context.Background()
		endpoint := ":9998"
		conn, err := serverGrpc.DialInsecure(
			ctx,
			serverGrpc.WithEndpoint(endpoint),
			serverGrpc.WithMiddleware(
				recovery.Recovery(),
			),
		)
		defer func() {
			if err != nil {
				if cerr := conn.Close(); cerr != nil {
					log.Infof("Failed to close conn to %s: %v", endpoint, cerr)
				}
				return
			}
			go func() {
				<-ctx.Done()
				if cerr := conn.Close(); cerr != nil {
					log.Infof("Failed to close conn to %s: %v", endpoint, cerr)
				}
			}()
		}()

		paralusJSON := NewParalusJSON()
		gateway := NewGateway(
			ctx,
			WithAddress(":9999"),
			WithConn(conn),
			WithServeMuxOptions(gwRuntime.WithMarshalerOption(jsonContentType, paralusJSON)),
			WithAnnotator(ParalusGatewayAnnotator),
			WithHandlers(helloworldpb.RegisterGreeterHandler),
		)
		err = gateway.Start(ctx)
		if err != nil {
			log.Fatalf("启动失败: %v", err)
		}

	}()
	<-stop
}

func TestGateway(t *testing.T) {
	stop := make(chan struct{})

	// go runServer(stop)
	// go runGateway(stop)
	go runGatewayV1(stop)

	defer func() {
		close(stop)
	}()

	client := http.Client{}
	resp, err := client.Get(fmt.Sprintf("http://localhost:9999/hello/%s", "world"))
	if err != nil {
		t.Error(err)
		return
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("body: %v", string(b))
	var obj helloworldpb.HelloReply
	err = json.Unmarshal(b, &obj)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("value: %v", obj.String())
}
