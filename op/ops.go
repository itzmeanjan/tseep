package op

type OP uint8

const (
	READ OP = iota + 1
	WRITE
)
