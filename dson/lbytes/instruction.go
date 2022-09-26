package lbytes

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
)

// ExecuteInstructions create the final value t with type T by
//
//   - Reading the instruction into a map, then
//   - Create JSON bytes from the map, and finally
//   - Read the JSON bytes into t
//
// In order to lessen the burden of manual mapping.
func ExecuteInstructions[T any](instructions []Instruction) (*T, error) {
	tMap := map[string]any{}
	for _, instruction := range instructions {
		value, err := instruction.ReadFunction()
		if err != nil {
			err := errors.Wrapf(err, `ExecuteInstructions error reading key "%v"`, instruction.Key)
			return nil, err
		}
		tMap[instruction.Key] = value
	}
	tBytes, err := json.Marshal(tMap)
	if err != nil {
		err := errors.Wrapf(err, `ExecuteInstructions error marshalling map "%v" to JSON`, tMap)
		return nil, err
	}

	var t T
	if err := json.Unmarshal(tBytes, &t); err != nil {
		err := errors.Wrapf(
			err, `ExecuteInstructions error unmarshalling bytes "%s" to type "%T"`,
			string(tBytes), t,
		)
		return nil, err
	}

	return &t, nil
}

func CreateNBytesReadFunction(reader *Reader, n int) func() (any, error) {
	return func() (any, error) {
		return reader.ReadBytes(n)
	}
}

func CreateIntReadFunction(reader *Reader) func() (any, error) {
	return func() (any, error) {
		return reader.ReadInt()
	}
}

func CreateStringReadFunction(reader *Reader, n int) func() (any, error) {
	return func() (any, error) {
		result, err := reader.ReadString(n)
		if err != nil {
			return "", err
		}
		// zero byte trimming is needed since that is how strings are laid out in a DSON file
		return strings.TrimRight(result, "\u0000"), nil
	}
}
