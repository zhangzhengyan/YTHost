package dataFrameEncoder

import (
	"encoding/binary"
	"io"
)

type Frame struct {
	Size int32
	Data []byte
}

type FrameDecoder struct {
	reader io.Reader
}
type FrameEncoder struct {
	writer io.Writer
}

func NewEncoder(writer io.Writer) *FrameEncoder {
	return &FrameEncoder{writer}
}
func NewDecoder(reader io.Reader) *FrameDecoder {
	return &FrameDecoder{reader}
}

func (mr *FrameDecoder) Decode() ([]byte, error) {
	var size int32
	if err := binary.Read(mr.reader, binary.BigEndian, &size); err != nil {
		return nil, err
	}
	var data = make([]byte, size)
	if err := binary.Read(mr.reader, binary.BigEndian, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (mw *FrameEncoder) Encode(data []byte) error {
	var size int32 = int32(len(data))
	if err := binary.Write(mw.writer, binary.BigEndian, size); err != nil {
		return err
	}
	if err := binary.Write(mw.writer, binary.BigEndian, data); err != nil {
		return err
	}
	return nil
}
