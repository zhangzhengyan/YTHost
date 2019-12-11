package yob

import (
	"encoding/binary"
	"io"
	"reflect"
)

type Encoder struct {
	io.Writer
}

func (ec *Encoder) Encode(indata interface{}) error {
	vl := reflect.ValueOf(indata)
	return binary.Write(ec.Writer, binary.BigEndian, vl.Bytes())
}

type Decoder struct {
	io.Reader
}

func (dc *Decoder) Decode(indata interface{}) error {
	vl := reflect.ValueOf(indata)
}
