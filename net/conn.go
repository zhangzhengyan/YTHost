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
	LocalPeer() peer.AddrInfo
}

type ytconn struct {
	manet.Conn
	remotePeer peer.AddrInfo
	localPeer  peer.AddrInfo
}

func (yc *ytconn) RemotePeer() peer.AddrInfo {
	return yc.remotePeer
}

func (yc *ytconn) LocalPeer() peer.AddrInfo {
	return yc.localPeer
}

// WarpConn 包装连接，交换peerinfo信息
func WarpConn(conn manet.Conn, pi peer.AddrInfo) (Conn, error) {
	done := make(chan struct{})
	errChan := make(chan error)

	ytp := YTP{
		ReadWriter: conn,
		LocalID:    pi.ID,
		LocaAddrs:  pi.Addrs,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := ytp.SendPeerInfo(ctx); err != nil {
		return nil, err
	}
	go func() {
		// 执行握手 交换peerInfo
		if err := ytp.Handshake(ctx); err != nil {
			errChan <- err
		} else {
			done <- struct{}{}
		}
	}()

	yc := new(ytconn)
	yc.Conn = conn
	yc.localPeer = peer.AddrInfo{ytp.LocalID, ytp.LocaAddrs}
	yc.remotePeer = peer.AddrInfo{ytp.RemoteID, ytp.RemoteAddrs}

	// 解除conn和ytp协议的绑定
	ytp.ReadWriter = nil

	select {
	case <-done:
		return yc, nil
	case err := <-errChan:
		return nil, err
	}
}
