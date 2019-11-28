package net

import (
	"encoding/gob"
	"github.com/graydream/YTHost/pbMsgHandler"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr-net"
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
	ec *gob.Encoder
	dc *gob.Decoder
}

func (yc *ytconn) RemotePeer() peer.AddrInfo {
	var pi peer.AddrInfo
	yc.dc.Decode(&pi)
	return pi
}

func (yc *ytconn) LocalPeer() peer.AddrInfo {
	return *yc.localPeer
}

// WarpConn 包装连接，交换peerinfo信息
func WarpConn(conn manet.Conn, pi *peer.AddrInfo) (Conn, error) {
	var yc = new(ytconn)

	yc.localPeer = pi
	yc.ec = gob.NewEncoder(conn)

	yc.ec.Encode(*pi)
	yc.dc = gob.NewDecoder(conn)

	return yc, nil
}
