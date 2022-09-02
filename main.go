package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"darkest-savior/dson"
	"darkest-savior/dson/lbytes"
)

func main() {
	// ui.Start()
	bytes, err := os.ReadFile("sample_data/persist.campaign_log.json")
	if err != nil {
		panic(err)
	}
	reader := lbytes.NewBytesReader(bytes)
	decodedFile, err := dson.DecodeDSON(reader)
	if err != nil {
		panic(err)
	}
	decodedFileBytes, err := json.Marshal(decodedFile)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("/tmp/dson-struct.json", decodedFileBytes, 0664)

	lhm := dson.ToLinkedHashMap(decodedFile.Fields)
	lhmBytes, err := lhm.MarshalJSON()
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("/tmp/dson.json", lhmBytes, 0664)
	if err != nil {
		panic(err)
	}
}
