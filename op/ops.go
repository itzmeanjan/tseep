package op

import (
	"encoding/binary"
	"io"
)

type OP uint8

const (
	READ     OP = iota + 1 // read request opcode
	WRITE                  // write request opcode
	RESPONSE               // response opcode
)

func (o OP) WriteTo(w io.Writer) (int64, error) {
	if err := binary.Write(w, binary.BigEndian, o); err != nil {
		return 0, err
	}

	return 1, nil
}

func (o *OP) ReadFrom(r io.Reader) (int64, error) {
	var op uint8

	if err := binary.Read(r, binary.BigEndian, &op); err != nil {
		return 0, err
	}

	*o = OP(op)
	return 1, nil
}

func ReadEnvelope(r io.Reader) (OP, uint32, error) {
	op := new(OP)

	if _, err := op.ReadFrom(r); err != nil {
		return 0, 0, err
	}

	var bodyLength uint32
	if err := binary.Read(r, binary.BigEndian, &bodyLength); err != nil {
		return *op, 0, err
	}

	return *op, bodyLength, nil
}
