// Package dson stores the code to decode and encode Darkest Dungeon JSON files.
package dson

import (
	"github.com/thanhnguyen2187/darkest-savior/dson/dheader"
)

func IsDSONFile(bs []byte) bool {
	return dheader.IsValidMagicNumber(bs[:4])
}
