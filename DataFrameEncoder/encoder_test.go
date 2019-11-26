package dataFrameEncoder

import (
	"bytes"
	"testing"
)

func TestMessageReader_Read(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	ec := FrameEncoder{buf}
	dc := FrameDecoder{buf}
	ec.Encode([]byte("1111"))
	if data, err := dc.Decode(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(string(data))
	}
	if _, err := dc.Decode(); err == nil {
		t.Fatal("连接应该已关闭")
	} else {
		t.Log(err)
	}
}
