package op

import "io"

type Key string

func (k *Key) len() int {
	return len(*k)
}

func (k *Key) writeTo(w io.Writer) (int64, error) {
	n, err := w.Write([]byte(*k))
	return int64(n), err
}

func (k *Key) readFrom(r io.Reader, n int64) (int64, error) {
	buf := make([]byte, n)
	if _, err := r.Read(buf); err != nil {
		return 0, err
	}

	*k = Key(buf)
	return n, nil
}
