package gateway

import (
	"bytes"
	"testing"

	"github.com/apus-run/gaea/internal/testdata/helloworld"
)

func TestParalusJSONMarshaller(t *testing.T) {
	m := NewParalusJSON()

	t1 := helloworld.HelloReply{
		Message: "hello",
	}

	yb, err := m.Marshal(&t1)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(string(yb))

	var t2 helloworld.HelloReply

	err = m.Unmarshal(yb, &t2)
	if err != nil {
		t.Error(err)
	}

	t.Log(t2)

	bb1 := new(bytes.Buffer)

	bb1.Write(yb)

	dec := m.NewDecoder(bb1)
	var t3 helloworld.HelloReply
	err = dec.Decode(&t3)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(t2)

	bb2 := new(bytes.Buffer)

	enc := m.NewEncoder(bb2)
	err = enc.Encode(&t1)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(bb2.String())
}
