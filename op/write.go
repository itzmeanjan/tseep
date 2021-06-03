package op

import (
	"encoding/binary"
	"io"
)

type WriteRequest struct {
	Key   *Key
	Value *Value
}

func (w *WriteRequest) Len() int {
	return w.Key.len() + w.Value.len()
}

func (w *WriteRequest) WriteEnvelope(wr io.Writer) (int64, error) {
	var total int64

	op := WRITE
	if _, err := op.WriteTo(wr); err != nil {
		return total, err
	}

	total += 1
	if err := binary.Write(wr, binary.BigEndian, uint16(w.Len()+2)); err != nil {
		return total, err
	}

	total += 2
	return total, nil
}

func (w *WriteRequest) WriteTo(wr io.Writer) (int64, error) {
	var total int64

	if err := binary.Write(wr, binary.BigEndian, uint8(w.Key.len())); err != nil {
		return total, err
	}

	total += 1
	n, err := w.Key.writeTo(wr)
	if err != nil {
		return total, err
	}

	total += n
	if err := binary.Write(wr, binary.BigEndian, uint8(w.Value.len())); err != nil {
		return total, err
	}

	total += 1
	n, err = w.Value.writeTo(wr)
	if err != nil {
		return total, err
	}

	total += n
	return total, nil
}

func (w *WriteRequest) ReadFrom(r io.Reader) (int64, error) {
	var total int64

	var keyLength uint8
	if err := binary.Read(r, binary.BigEndian, &keyLength); err != nil {
		return total, err
	}

	total += 1
	key := new(Key)
	n, err := key.readFrom(r, int64(keyLength))
	if err != nil {
		return total, err
	}

	total += n
	var valLength uint8
	if err := binary.Read(r, binary.BigEndian, &valLength); err != nil {
		return total, err
	}

	total += 1
	val := new(Value)
	n, err = val.readFrom(r, int64(valLength))
	if err != nil {
		return total, err
	}

	total += n
	w.Key = key
	w.Value = val
	return total, nil
}
