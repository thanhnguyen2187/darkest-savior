package main

import (
	"os"

	"darkest-savior/dson"
)

func main() {
	bytes, err := os.ReadFile("sample-data/persist.town.json")
	if err != nil {
		panic(err)
	}
	reader := dson.NewBytesReader(bytes)
	decodedFile, err := dson.DecodeDSON(&reader)
	if err != nil {
		panic(err)
	}
	println(decodedFile.Header.MagicNumber)
	println(decodedFile.Meta1Blocks)
}
