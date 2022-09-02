package dhash

import (
	_ "embed"
	"strings"

	"github.com/samber/lo"
)

//go:embed names.txt
var names string

var NameByHash map[int32]string

func init() {
	namesSlice := strings.Split(names, "\n")
	namesSlice = lo.Filter(
		namesSlice,
		func(line string, _ int) bool {
			return len(strings.TrimSpace(line)) > 0
		},
	)

	NameByHash = lo.SliceToMap[string, int32, string](
		namesSlice,
		func(name string) (int32, string) {
			return HashString(name), "###" + name
		},
	)
}
