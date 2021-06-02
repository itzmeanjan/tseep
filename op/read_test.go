package op_test

import (
	"bytes"
	"testing"

	"github.com/itzmeanjan/tseep/op"
)

func TestReadRequest(t *testing.T) {
	key := op.Key("hello")
	readReq1 := op.ReadRequest{
		Key: &key,
	}
	stream := new(bytes.Buffer)

	if _, err := readReq1.WriteTo(stream); err != nil {
		t.Fatalf("Failed to write : %s\n", err.Error())
	}

	readReq2 := new(op.ReadRequest)
	if _, err := readReq2.ReadFrom(stream); err != nil {
		t.Fatalf("Failed to read : %s\n", err.Error())
	}

	if string(*readReq1.Key) != string(*readReq2.Key) {
		t.Fatalf("Bad write to/ read from stream\n")
	}
}
