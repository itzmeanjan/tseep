package op

import "io"

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
