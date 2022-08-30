package ds

import (
	"encoding/json"
	"fmt"
)

func JSONDumps[T any](t T) string {
	tBytes, err := json.Marshal(t)
	if err != nil {
		return fmt.Errorf("JSONDumps error %w", err).Error()
	}

	return string(tBytes)
}
