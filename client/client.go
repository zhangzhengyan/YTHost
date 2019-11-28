package client

import (
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"

	"net/rpc"
)

type YTHostClient struct {
	*rpc.Client
	localPeer *peer.AddrInfo
}

func (yc *YTHostClient) RemotePeer() peer.AddrInfo {
	var pi peer.AddrInfo

	if err := yc.Call("as.RemotePeerInfo", "", &pi); err != nil {
		fmt.Println(err)
	}

	return pi
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
