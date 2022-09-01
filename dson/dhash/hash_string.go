package dhash

import (
	"github.com/samber/lo"
)

func HashString(s string) int32 {
	return lo.Reduce(
		[]byte(s),
		func(result int32, b byte, _ int) int32 {
			return result*53 + int32(b)
		},
		0,
	)
}
