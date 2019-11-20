package message

import (
	"bytes"
	"fmt"
	"testing"
)

func TestMessageReader_Read(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	mw :=MessageWriter{buf}
	mr :=MessageReader{buf}
	mw.Write([]byte("这是一段测试文本!!"))
	var res = make([]byte,26)
	mr.Read(res)
	for _,v:=range res{
		fmt.Println(v)
	}
}
