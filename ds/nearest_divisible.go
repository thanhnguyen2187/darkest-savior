package ds

import (
	"fmt"
)

func NearestDivisibleByM(n int, m int) int {
	for i := n; i < n+m; i++ {
		if i%m == 0 {
			return i
		}
	}

	err := fmt.Errorf(
		`NearestDivisibleByM unreachable code with n = %d and m = %d`,
		n, m,
	)
	panic(err)
}
