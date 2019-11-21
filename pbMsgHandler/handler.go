package pbMsgHandler

import (
	"github.com/graydream/YTHost/YTHostError"
	mnet "github.com/multiformats/go-multiaddr-net"
)

type HandlerFunc func(msgId uint16, data []byte,conn mnet.Conn)

type PBMsgHanderMap struct {
	handlerMap map[string]map[uint16]HandlerFunc
}

func (pm *PBMsgHanderMap)RegisterMsgHandler(protocol string,msgID uint16,handler HandlerFunc) *YTHostError.YTError {
	if pm.handlerMap == nil {
		pm.handlerMap = make(map[string]map[uint16]HandlerFunc)
	}
	if _,ok:=pm.handlerMap[protocol];!ok {
		pm.handlerMap[protocol] = make(map[uint16]HandlerFunc)
	}
	if _,ok:=pm.handlerMap[protocol][msgID];ok{
		return YTHostError.NewError(0, "handler already exist")
	}
	pm.handlerMap[protocol][msgID] = handler
	return nil
}

func (pm *PBMsgHanderMap)RemoveMsgHandler(protocol string,msgID uint16) {
	if pm.handlerMap == nil {
		pm.handlerMap = make(map[string]map[uint16]HandlerFunc)
	}
	if pm.handlerMap[protocol]==nil {
		pm.handlerMap[protocol] = make(map[uint16]HandlerFunc)
	}

	delete(pm.handlerMap[protocol],msgID)
}
