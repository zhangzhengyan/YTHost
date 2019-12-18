package service

import (
	"fmt"
	"github.com/graydream/YTHost/encrypt"
	"github.com/graydream/YTHost/peerInfo"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/mr-tron/base58"
	"github.com/multiformats/go-multiaddr"
	"github.com/yottachain/YTCrypto"
	"sync"
)

type MsgId int32

type Handler func(requestData []byte, head Head) ([]byte, error)

type HandlerMap map[int32]Handler

type Head struct {
	MsgId        int32
	RemotePeerID peer.ID
	RemoteAddrs  []multiaddr.Multiaddr
	RemotePubKey []byte
}

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
func (hm HandlerMap) RemoveHandler(id int32) {
	delete(hm, id)
}
func (hm HandlerMap) RemoveGlobalHandler() {
	delete(hm, 0x0)
}


type MsgService struct {
	Handler HandlerMap
	Pi      peerInfo.PeerInfo
	LocalPriKey crypto.PrivKey
	sync.Map
}

type Request struct {
	MsgId          int32
	ReqData        []byte
	RemotePeerInfo PeerInfo
}

type Response struct {
	Data []byte
}

type IdToMsgPriKey struct {
	ID     peer.ID
	MsgPriKey		[]byte
}

func (ms *MsgService) Ping(req string, res *string) error {
	*res = "pong"
	return nil
}

func (ms *MsgService) RegisterMsgPriKey(clinetMsgKey IdToMsgPriKey, res *bool) error {
	hostPriKey, _ := ms.LocalPriKey.Raw()

	peerId := clinetMsgKey.ID
	msgPriKey := clinetMsgKey.MsgPriKey

	/*pr, err := base58.Decode(base58.Encode(hostPriKey))
	if err == nil {
		pr = append([]byte{0x80}, pr[0:32]...)
	} else {
		return err
	}*/

	pr := append([]byte{0x80}, hostPriKey[0:32]...)

	finalMsgPriKey, err:= YTCrypto.ECCDecrypt(msgPriKey, base58.Encode(pr))
	if err != nil {
		*res = false
		return err
	}

	//fmt.Printf("server recive client gen prikey is %s\n", base58.Encode(finalMsgPriKey))

	ms.Map.Store(peerId, finalMsgPriKey)
	*res = true

	return nil
}

func (ms *MsgService) HandleMsg(req Request, data *Response) error {

	if ms.Handler == nil {
		return fmt.Errorf("no handler %x", req.MsgId)
	}

	// 0x0～0x10 为保留全局消息处理器
	h, ok := ms.Handler[0x0]
	// 如果没有全局处理器，调用局部处理器
	if !ok {
		h, ok = ms.Handler[req.MsgId]
	}
	head := Head{}
	head.MsgId = req.MsgId
	head.RemotePeerID = req.RemotePeerInfo.ID
	head.RemotePubKey = req.RemotePeerInfo.PubKey

	for _, v := range req.RemotePeerInfo.Addrs {
		ma, _ := multiaddr.NewMultiaddr(v)
		head.RemoteAddrs = append(head.RemoteAddrs, ma)
	}

	_k, ok := ms.Map.Load(head.RemotePeerID)
	if !ok {
		return fmt.Errorf("no msgPriKey %x", head.RemotePeerID)
	}
	msgKey := _k.([]byte)

	//fmt.Printf("before at aes decode msg is %s\n", string(req.ReqData))

	reqData, err := encrypt.AesCBCDncrypt(req.ReqData, msgKey)
	if err != nil {
		return fmt.Errorf("AesCBCDncrypt msg error %x", head.RemotePeerID)
	}

	//fmt.Printf("secret key [%s] ---------> after at aes decode msg is [%s]\n", base58.Encode(msgKey), string(reqData))

	if ok {
		if resdata, err := h(reqData, head); err != nil {
			return nil
		} else {
			data.Data = resdata
			return nil
		}
	} else {
		return fmt.Errorf("no handler %x", req.MsgId)
	}
}
