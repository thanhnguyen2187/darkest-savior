package lbytes

import (
	"bytes"
)

type (
	Reader struct {
		bytes.Reader
	}
	Instruction struct {
		Key          string
		ReadFunction ReadFunction
	}
	ReadFunction func() (any, error)
)
