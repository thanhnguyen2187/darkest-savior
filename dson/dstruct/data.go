package dstruct

import (
	"github.com/thanhnguyen2187/darkest-savior/dson/dfield"
	"github.com/thanhnguyen2187/darkest-savior/dson/dheader"
	"github.com/thanhnguyen2187/darkest-savior/dson/dmeta1"
	"github.com/thanhnguyen2187/darkest-savior/dson/dmeta2"
)

type (
	Struct struct {
		Header     dheader.Header     `json:"header"`
		Meta1Block []dmeta1.Entry     `json:"meta_1_block"`
		Meta2Block []dmeta2.Entry     `json:"meta_2_block"`
		Fields     []dfield.DataField `json:"fields"`
	}
)
