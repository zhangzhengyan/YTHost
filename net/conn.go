package net

import (
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr-net"
	"net/rpc"
)

type Conn interface {
	manet.Conn
	RemotePeer() *peer.AddrInfo
	LocalPeer() *peer.AddrInfo
}

type ytconn struct {
	manet.Conn
	client *rpc.Client
	srv    *rpc.Server
	RpcService
}

func (yc *ytconn) RemotePeer() *peer.AddrInfo {
	pi := &peer.AddrInfo{}
	err := yc.client.Call("RpcService.PeerInfo", "", pi)
	if err != nil {
		fmt.Println(err)
	}
	return pi
}

func (yc *ytconn) LocalPeer() *peer.AddrInfo {
	return yc.localPeer
}

type RpcService struct {
	remotePeer *peer.AddrInfo
	localPeer  *peer.AddrInfo
}

func (rs *RpcService) PeerInfo(request string, reply *peer.AddrInfo) error {
	reply.ID = rs.localPeer.ID
	reply.Addrs = rs.localPeer.Addrs
	fmt.Println(reply)
	return nil
}

// WarpConn 包装连接，交换peerinfo信息
func WarpConn(conn manet.Conn, pi *peer.AddrInfo) (Conn, error) {
	var yc = new(ytconn)

	rs := RpcService{localPeer: pi}

	srv := rpc.NewServer()
	yc.srv = srv
	err := srv.Register(&rs)
	if err != nil {
		return nil, err
	}
	go srv.ServeConn(conn)

	client := rpc.NewClient(conn)
	yc.client = client
	yc.RpcService = rs

	return yc, nil
}
