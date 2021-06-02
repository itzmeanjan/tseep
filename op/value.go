package op

import (
	"encoding/binary"
	"io"
)

type Value []byte

func (v *Value) len() int {
	return len(*v)
}

func (v *Value) writeTo(w io.Writer) (int64, error) {
	n, err := w.Write(*v)
	return int64(n), err
}

func (v *Value) readFrom(r io.Reader, n int64) (int64, error) {
	buf := make([]byte, n)
	if _, err := r.Read(buf); err != nil {
		return 0, err
	}

	*v = buf
	return n, nil
}

func (v *Value) WriteTo(w io.Writer) (int64, error) {
	var total int64

	op := RESPONSE
	if _, err := op.WriteTo(w); err != nil {
		return total, err
	}

	total += 1
	if err := binary.Write(w, binary.BigEndian, uint8(v.len())); err != nil {
		return total, err
	}

	total += 1
	n, err := v.writeTo(w)
	if err != nil {
		return total, err
	}

	total += n
	return total, nil
}
