package op

import (
	"encoding/binary"
	"io"
)

type ReadRequest struct {
	Key *Key
}

func (r *ReadRequest) Len() int {
	return r.Key.len()
}

func (r *ReadRequest) WriteEnvelope(w io.Writer) (int64, error) {
	var total int64

	op := READ
	if _, err := op.WriteTo(w); err != nil {
		return total, err
	}

	total += 1
	if err := binary.Write(w, binary.BigEndian, uint32(r.Len()+1)); err != nil {
		return total, err
	}

	total += 4
	return total, nil
}

func (r *ReadRequest) WriteTo(w io.Writer) (int64, error) {
	var total int64

	if err := binary.Write(w, binary.BigEndian, uint8(r.Key.len())); err != nil {
		return total, err
	}

	total += 1
	n, err := r.Key.writeTo(w)
	if err != nil {
		return total, err
	}

	total += n
	return total, nil
}

func (r *ReadRequest) ReadFrom(rd io.Reader) (int64, error) {
	var (
		total   int64
		keySize uint8
	)

	if err := binary.Read(rd, binary.BigEndian, &keySize); err != nil {
		return total, err
	}

	total += 1
	key := new(Key)
	n, err := key.readFrom(rd, int64(keySize))
	if err != nil {
		return total, err
	}

	total += n
	r.Key = key
	return total, nil
}
