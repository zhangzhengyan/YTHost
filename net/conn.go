package net

import (
	"context"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr-net"
	"time"
)

type Conn interface {
	manet.Conn
	RemotePeer() peer.AddrInfo
}

type ytconn struct {
	manet.Conn
	remotePeer peer.AddrInfo
}

func (yc *ytconn) RemotePeer() peer.AddrInfo {
	return yc.remotePeer
}

func WarpConn(conn manet.Conn, pi peer.AddrInfo) (Conn, error) {

	ytp := YTP{
		ReadWriter: conn,
		LocalID:    pi.ID,
		LocaAddrs:  pi.Addrs,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := ytp.Handshake(ctx); err != nil {
		return nil, err
	}

	return nil, nil
}

//func WarpConn(conn manet.Conn) Conn{
//	yc := ytconn{
//		conn,
//	}
//	return yc
//}
