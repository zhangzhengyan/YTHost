package message

import (
	"encoding/binary"
	"io"
)

type Message struct {
	Size int32
	Data []byte
}

type MessageReader struct {
	buf io.Reader
}
type MessageWriter struct {
	buf io.Writer
}

func (mr *MessageReader)Read([]byte)(int32,error) {
	var size int32
	if err := binary.Read(mr.buf,binary.BigEndian, &size);err != nil {
		return 0,err
	}
	var data = make([]byte, size)
	if err :=binary.Read(mr.buf, binary.BigEndian, &data);err != nil {
		return 0,err
	}
	return size,nil
}

func (mw *MessageWriter)Write(data []byte)(error) {
	var size int32 = int32(len(data))
	if err:=binary.Write(mw.buf,binary.BigEndian,size);err != nil {
		return err
	}
	if err:=binary.Write(mw.buf,binary.BigEndian,data);err != nil {
		return err
	}
	return nil
}