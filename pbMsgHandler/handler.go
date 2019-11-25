package pbMsgHandler

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/graydream/YTHost/DataFrameEncoder"
	"github.com/graydream/YTHost/YTHostError"
	mnet "github.com/multiformats/go-multiaddr-net"
)
// HandlerFunc 消息处理函数
type HandlerFunc func(msgId uint16, data []byte, handler *PBMsgHandler)

type Msg struct {
	ID uint16
	Data []byte
}

// PBMsgHanderMap 消息处理函数映射
type PBMsgHanderMap struct {
	handlerMap map[uint16]HandlerFunc
}
// RegisterMsgHandler 注册消息处理函数
func (pm *PBMsgHanderMap)RegisterMsgHandler(msgID uint16,handler HandlerFunc) *YTHostError.YTError {
	if pm.handlerMap == nil {
		pm.handlerMap = make(map[uint16]HandlerFunc)
	}
	if _,ok:=pm.handlerMap[msgID];ok{
		return YTHostError.NewError(0, "handler already exist")
	}
	pm.handlerMap[msgID] = handler
	return nil
}
// RemoveMsgHandler 移除消息处理函数
func (pm *PBMsgHanderMap)RemoveMsgHandler(msgID uint16) {
	if pm.handlerMap == nil {
		pm.handlerMap = make(map[uint16]HandlerFunc)
	}

	delete(pm.handlerMap,msgID)
}

func (pm *PBMsgHanderMap)Call(msgID uint16,data []byte,handler *PBMsgHandler) error {
	if pm.handlerMap == nil {
		pm.handlerMap = make(map[uint16]HandlerFunc)
	}
	if hf,ok:=pm.handlerMap[msgID];ok{
		hf(msgID,data,handler)
	} else {
		return fmt.Errorf("No msgId message")
	}
	return nil
}

// PBMsgHandler 消息处理器
type PBMsgHandler struct {
	mnet.Conn
	ec *dataFrameEncoder.FrameEncoder
	dc *dataFrameEncoder.FrameDecoder
}

func NewPBMsgHander(conn mnet.Conn) *PBMsgHandler {
	pbmh:=new(PBMsgHandler)
	pbmh.Conn = conn
	pbmh.ec = dataFrameEncoder.NewEncoder(conn)
	pbmh.dc = dataFrameEncoder.NewDecoder(conn)
	return pbmh
}

// SendMsg 发送消息
func (pbmh *PBMsgHandler)SendMsg(msgId uint16,msg proto.Message) error {
	buf:=bytes.NewBuffer([]byte{})
	data,err:=proto.Marshal(msg)
	if err != nil {
		return err
	}
	err=binary.Write(buf,binary.BigEndian,msgId)
	if err != nil {
		return err
	}
	err=binary.Write(buf,binary.BigEndian,data)
	if err != nil {
		return err
	}
	return pbmh.ec.Encode(buf.Bytes())
}

// SendMsg 发送消息
func (pbmh *PBMsgHandler)SendDataMsg(msgId uint16,data []byte) error {
	buf:=bytes.NewBuffer([]byte{})

	err:=binary.Write(buf,binary.BigEndian,msgId)
	if err != nil {
		return err
	}
	err=binary.Write(buf,binary.BigEndian,data)
	if err != nil {
		return err
	}
	return pbmh.ec.Encode(buf.Bytes())
}

func (pbmh *PBMsgHandler)Accept() (*Msg,error) {
	data,err:=pbmh.dc.Decode()
	if err !=nil {
		return nil,err
	}
	buf:=bytes.NewBuffer(data)
	var msgId uint16
	if err:=binary.Read(buf,binary.BigEndian,&msgId);err != nil {
		return nil,err
	}
	return &Msg{msgId,data[2:]},nil
}

