package pbMsgHandler

import (
	"bytes"
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"github.com/graydream/YTHost/DataFrameEncoder"
	"github.com/graydream/YTHost/YTHostError"
	mnet "github.com/multiformats/go-multiaddr-net"
)
// HandlerFunc 消息处理函数
type HandlerFunc func(msgId uint16, data []byte, handler *PBMsgHandler)

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
func (pm *PBMsgHanderMap)RemoveMsgHandler(protocol string,msgID uint16) {
	if pm.handlerMap == nil {
		pm.handlerMap = make(map[uint16]HandlerFunc)
	}

	delete(pm.handlerMap,msgID)
}

// PBMsgHandler 消息处理器
type PBMsgHandler struct {
	mnet.Conn
	ec *dataFrameEncoder.FrameEncoder
	dc *dataFrameEncoder.FrameDecoder
	// 消息处理函数表
	PBMsgHanderMap
	//// 中间件
	//middleware []middleware.Middleware
}

func NewPBMsgHander(conn mnet.Conn) *PBMsgHandler {
	pbmh:=new(PBMsgHandler)
	pbmh.Conn = conn
	pbmh.ec = dataFrameEncoder.NewEncoder(conn)
	pbmh.dc = dataFrameEncoder.NewDecoder(conn)
	go pbmh.serve()
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

func (pbmh *PBMsgHandler)serve() error {
	for {
		data,err:=pbmh.dc.Decode()
		if err !=nil {
			return err
		}
		buf:=bytes.NewBuffer(data)
		var msgId uint16
		if err:=binary.Read(buf,binary.BigEndian,&msgId);err != nil {
			return err
		}
		if handler,ok:=pbmh.handlerMap[msgId];ok{
			handler(msgId,data[2:],pbmh)
		}
	}
}

