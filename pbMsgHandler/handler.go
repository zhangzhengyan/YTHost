package pbMsgHandler

import (
	"bytes"
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"github.com/graydream/YTHost/DataFrameEncoder"
	"github.com/graydream/YTHost/YTHostError"
	mnet "github.com/multiformats/go-multiaddr-net"
)

type HandlerFunc func(msgId uint16, data []byte,conn mnet.Conn)

type PBMsgHanderMap struct {
	handlerMap map[uint16]HandlerFunc
}

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

func (pm *PBMsgHanderMap)RemoveMsgHandler(protocol string,msgID uint16) {
	if pm.handlerMap == nil {
		pm.handlerMap = make(map[uint16]HandlerFunc)
	}

	delete(pm.handlerMap,msgID)
}

type PBMsgHander struct {
	conn mnet.Conn
	ec *dataFrameEncoder.FrameEncoder
	dc *dataFrameEncoder.FrameDecoder
	// 消息处理函数表
	PBMsgHanderMap
	//// 中间件
	//middleware []middleware.Middleware
}

func NewPBMsgHander(conn mnet.Conn) *PBMsgHander {
	pbmh:=new(PBMsgHander)
	pbmh.ec = dataFrameEncoder.NewEncoder(conn)
	pbmh.dc = dataFrameEncoder.NewDecoder(conn)
	return pbmh
}

func (pbmh *PBMsgHander)SendMsg(msgId uint16,msg proto.Message) error {
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

func (pbmh *PBMsgHander)Serve() error {
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
		pbmh.handlerMap[msgId](msgId,data[2:],pbmh.conn)
	}
}

