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

	if _, err := readReq1.WriteEnvelope(stream); err != nil {
		t.Fatalf("Failed to write envelope : %s\n", err.Error())
	}

	if _, err := readReq1.WriteTo(stream); err != nil {
		t.Fatalf("Failed to write : %s\n", err.Error())
	}

	readReq2 := new(op.ReadRequest)

	opcode, bodyLen, err := op.ReadEnvelope(stream)
	if err != nil {
		t.Fatalf("Failed to read envelope : %s\n", err.Error())
	}

	if opcode != op.READ {
		t.Fatalf("Expected READ opcode\n")
	}

	if _, err := readReq2.ReadFrom(stream); err != nil {
		t.Fatalf("Failed to read : %s\n", err.Error())
	}

	if int(bodyLen) != readReq2.Len()+1 {
		t.Fatalf("Bad length denotation in envelope\n")
	}

	if string(*readReq1.Key) != string(*readReq2.Key) {
		t.Fatalf("Bad write to/ read from stream\n")
	}
}
