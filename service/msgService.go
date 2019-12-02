package service

import "fmt"

type MsgId uint16

type Handler func(requestData []byte) ([]byte, error)

type HandlerMap map[MsgId]Handler

func (hm HandlerMap) RegisterHandler(id MsgId, handlerFunc Handler) {
	if hm == nil {
		hm = make(HandlerMap)
	}
	hm[id] = handlerFunc
}

func (hm HandlerMap) RemoveHandler(id MsgId) {
	delete(hm, id)
}

type MsgService struct {
	Handler HandlerMap
}

type Request struct {
	MsgId   MsgId
	ReqData []byte
}

type Response struct {
	Data []byte
}

func (ms *MsgService) HandleMsg(req Request, data *Response) error {
	if ms.Handler == nil {
		return fmt.Errorf("no handler")
	}
	if h, ok := ms.Handler[req.MsgId]; ok {
		if resdata, err := h(req.ReqData); err != nil {
			return nil
		} else {
			data.Data = resdata
			return nil
		}
	} else {
		return fmt.Errorf("no handler")
	}
}
