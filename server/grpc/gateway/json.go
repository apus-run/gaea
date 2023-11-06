package gateway

import (
	"io"

	"github.com/bytedance/sonic"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

const (
	jsonContentType string = "application/json"
)

var sonicAPI = sonic.Config{
	EscapeHTML:       true, // 安全需求
	CompactMarshaler: true, // 兼容需求
}.Froze()

// paralusJSON is the paralus object to json marshaller
type paralusJSON struct {
}

// NewParalusJSON returns new grpc gateway paralus json marshaller
func NewParalusJSON() runtime.Marshaler {
	return &paralusJSON{}
}

// ContentType returns the Content-Type which this marshaler is responsible for.
func (m *paralusJSON) ContentType(_ interface{}) string {
	return jsonContentType
}

// Marshal marshals "v" into byte sequence.
func (m *paralusJSON) Marshal(v interface{}) ([]byte, error) {
	return sonic.Marshal(v)
}

// Unmarshal unmarshals "data" into "v".
// "v" must be a pointer value.
func (m *paralusJSON) Unmarshal(data []byte, v interface{}) error {
	return sonic.Unmarshal(data, v)
}

// NewDecoder returns a Decoder which reads byte sequence from "r".
func (m *paralusJSON) NewDecoder(r io.Reader) runtime.Decoder {
	return sonicAPI.NewDecoder(r)
}

// NewEncoder returns an Encoder which writes bytes sequence into "w".
func (m *paralusJSON) NewEncoder(w io.Writer) runtime.Encoder {
	return sonicAPI.NewEncoder(w)
}

// Delimiter for newline encoded JSON streams.
func (m *paralusJSON) Delimiter() []byte {
	return []byte("\n")
}
