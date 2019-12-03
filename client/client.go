package client

import (
	"context"
	"fmt"
	"github.com/graydream/YTHost/service"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"

	"net/rpc"
)

type YTHostClient struct {
	*rpc.Client
	localPeer *peer.AddrInfo
}

func (yc *YTHostClient) RemotePeer() peer.AddrInfo {
	var pi service.PeerInfo
	var ai peer.AddrInfo

	if err := yc.Call("as.RemotePeerInfo", "", &pi); err != nil {
		fmt.Println(err)
	}
	ai.ID = pi.ID
	for _, addr := range pi.Addrs {
		ma, _ := multiaddr.NewMultiaddr(addr)
		ai.Addrs = append(ai.Addrs, ma)
	}

	return ai
}

func (yc *YTHostClient) LocalPeer() peer.AddrInfo {
	return *yc.localPeer
}

func WarpClient(clt *rpc.Client, pi *peer.AddrInfo) (*YTHostClient, error) {
	var yc = new(YTHostClient)
	yc.Client = clt
	yc.localPeer = pi
	return yc, nil
}

func (yc *YTHostClient) SendMsg(ctx context.Context, id int32, data []byte) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("ctx time out")
	default:
		var res service.Response
		if err := yc.Call("ms.HandleMsg", service.Request{id, data}, &res); err != nil {
			return nil, err
		} else {
			return res.Data, nil
		}
	}
}
