package DataFrameEncoder

import (
	"bytes"
	"fmt"
	"testing"
)

func TestMessageReader_Read(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	ec:= FrameEncoder{buf}
	dc:= FrameDecoder{buf}
	ec.Encode([]byte("1111"))
	ec.Encode([]byte("99999"))
	ec.Encode([]byte("test"))

	fmt.Println(string(dc.Decode()))
	fmt.Println(string(dc.Decode()))
	fmt.Println(string(dc.Decode()))
}
