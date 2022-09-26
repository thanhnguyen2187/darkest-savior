package dstruct

import (
	"darkest-savior/dson/dfield"
	"darkest-savior/dson/dheader"
	"darkest-savior/dson/dmeta1"
	"darkest-savior/dson/dmeta2"
)

type (
	Struct struct {
		Header     dheader.Header     `json:"header"`
		Meta1Block []dmeta1.Entry     `json:"meta_1_block"`
		Meta2Block []dmeta2.Entry     `json:"meta_2_block"`
		Fields     []dfield.DataField `json:"fields"`
	}
)
