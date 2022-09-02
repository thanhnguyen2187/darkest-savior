package main

import (
	"os"

	"darkest-savior/dson"
	"darkest-savior/dson/lbytes"
)

func main() {
	// ui.Start()
	bytes, err := os.ReadFile("sample-data/persist.town.json")
	if err != nil {
		panic(err)
	}
	reader := lbytes.NewBytesReader(bytes)
	decodedFile, err := dson.DecodeDSON(&reader)
	if err != nil {
		panic(err)
	}
	println(decodedFile.Header.MagicNumber)
	println(decodedFile.Meta1Blocks)
	lhm := dson.ToLinkedHashMap(decodedFile.Fields)
	lhmBytes, err := lhm.MarshalJSON()
	if err != nil {
		panic(err)
	}
	println(string(lhmBytes))
}
