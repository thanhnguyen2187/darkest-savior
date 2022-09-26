package ds

import (
	"fmt"
)

type (
	ErrUnreachableCode struct {
		Caller string
	}
)

func (r ErrUnreachableCode) Error() string {
	return fmt.Sprintf("%s: unreachable code", r.Caller)
}
