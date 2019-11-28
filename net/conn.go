package net

import (
	"fmt"
	"github.com/graydream/YTHost/pbMsgHandler"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr-net"
	"net/rpc"
)

type Conn interface {
	manet.Conn
	RemotePeer() peer.AddrInfo
	LocalPeer() peer.AddrInfo
}

type ytconn struct {
	manet.Conn
	localPeer  *peer.AddrInfo
	remotePeer *peer.AddrInfo
	*pbMsgHandler.PBMsgHandler
	srv *rpc.Server
	clt *rpc.Client
}

func (yc *ytconn) RemotePeer() peer.AddrInfo {
	var pi peer.AddrInfo
	if err := yc.clt.Call("yc.PeerInfo", "111", &pi); err != nil {
		fmt.Println(err)
	}
	return pi
}

func (yc *ytconn) LocalPeer() peer.AddrInfo {
	return *yc.localPeer
}

type rpcService struct {
}

func (yc *ytconn) PeerInfo(req string, rep *peer.AddrInfo) error {
	rep.ID = yc.localPeer.ID
	rep.Addrs = yc.localPeer.Addrs
	return nil
}

// WarpConn 包装连接，交换peerinfo信息
func WarpConn(conn manet.Conn, pi *peer.AddrInfo) (Conn, error) {
	var yc = new(ytconn)

	yc.Conn = conn
	yc.localPeer = pi

	yc.srv = rpc.NewServer()
	if err := yc.srv.RegisterName("yc", yc); err != nil {
		return nil, err
	}
	go yc.srv.ServeConn(yc.Conn)

	yc.clt = rpc.NewClient(yc)
	return yc, nil
}
