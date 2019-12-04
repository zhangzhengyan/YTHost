package service

import (
	"fmt"
	"github.com/graydream/YTHost/peerInfo"
)

type MsgId int32

type Handler func(requestData []byte, remotePeer peerInfo.PeerInfo) ([]byte, error)

type HandlerMap map[int32]Handler

// RegisterHandler 注册消息处理器
func (hm HandlerMap) RegisterHandler(id int32, handlerFunc Handler) error {
	if id < 0x10 {
		return fmt.Errorf("msgID need >= 0x10")
	}
	hm.registerHandler(id, handlerFunc)
	return nil
}

func (hm HandlerMap) registerHandler(id int32, handlerFunc Handler) {
	if hm == nil {
		hm = make(HandlerMap)
	}
	hm[id] = handlerFunc
}

// RegisterGlobalMsgHandler 注册全局消息处理器
func (hm HandlerMap) RegisterGlobalMsgHandler(handlerFunc Handler) {
	hm.registerHandler(0x0, handlerFunc)
}

// RemoveHandler 移除消息处理器
func (hm HandlerMap) RemoveMsgHandler(id int32) {
	delete(hm, id)
}
func (hm HandlerMap) RemoveGlobalMsgHandler() {
	delete(hm, 0x0)
}

type MsgService struct {
	Handler HandlerMap
	Pi      peerInfo.PeerInfo
}

type Request struct {
	MsgId   int32
	ReqData []byte
}

type Response struct {
	Data []byte
}

func (ms *MsgService) HandleMsg(req Request, data *Response) error {

	if ms.Handler == nil {
		return fmt.Errorf("no handler")
	}

	// 0x0～0x10 为保留全局消息处理器
	h, ok := ms.Handler[0x0]
	// 如果没有全局处理器，调用局部处理器
	if !ok {
		h, ok = ms.Handler[req.MsgId]
	}
	if ok {
		if resdata, err := h(req.ReqData, ms.Pi); err != nil {
			return nil
		} else {
			data.Data = resdata
			return nil
		}
	} else {
		return fmt.Errorf("no handler")
	}
}
