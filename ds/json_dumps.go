package ds

import (
	"encoding/json"
	"fmt"
)

func DumpJSON[T any](t T) string {
	tBytes, err := json.Marshal(t)
	if err != nil {
		return fmt.Errorf("DumpJSON error %w", err).Error()
	}

	return string(tBytes)
}
