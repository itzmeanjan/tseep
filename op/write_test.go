package op_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/itzmeanjan/tseep/op"
)

func TestWriteRequest(t *testing.T) {
	key := op.Key("hello")
	val := op.Value("world")
	writeRequest1 := op.WriteRequest{Key: &key, Value: &val}
	stream := new(bytes.Buffer)

	if _, err := writeRequest1.WriteTo(stream); err != nil {
		t.Fatalf("Failed to write to stream : %s\n", err.Error())
	}

	writeRequest2 := new(op.WriteRequest)
	if _, err := writeRequest2.ReadFrom(stream); err != nil {
		t.Fatalf("Failed to read from stream : %s\n", err.Error())
	}

	if strings.Compare(string(*writeRequest1.Key), string(*writeRequest2.Key)) != 0 {
		t.Fatalf("[Key] Bad read to/ write from stream\n")
	}

	if !bytes.Equal(*writeRequest1.Value, *writeRequest2.Value) {
		t.Fatalf("[Value] Bad read to/ write from stream\n")
	}
}
