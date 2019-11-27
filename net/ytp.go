package net

import (
	"context"
	"encoding/gob"
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"io"
)

// YTP
// 此处为文本协议
// ID：表示peerID
// Addr: 表示maddr地址如果有多个地址则用多行表示
// Pubkey: 表示加密公钥，有则进行加密传输，无则不进行加密，加密传输使用对称加密，用公钥加密加密私钥后传输
// |end| 是终止标志
type YTP struct {
	io.ReadWriter
	LocalID        peer.ID
	LocaAddrs      []multiaddr.Multiaddr
	RemoteIsYTHost bool
	RemoteID       peer.ID
	RemoteAddrs    []multiaddr.Multiaddr
	Finish         bool
}

func (ytp *YTP) Handshake(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("ctx time out")
	default:
		dc := gob.NewDecoder(ytp)
		var pi peer.AddrInfo
		if err := dc.Decode(&pi); err != nil {
			return err
		}
		ytp.RemoteID = pi.ID
		ytp.RemoteAddrs = pi.Addrs
		return nil
	}
}

//func (ytp *YTP) confirmYThost(ctx context.Context) error {
//}

func (ytp *YTP) SendPeerInfo(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("ctx time out")
	default:
		ec := gob.NewEncoder(ytp.ReadWriter)

		var pi peer.AddrInfo
		pi.ID = ytp.LocalID
		pi.Addrs = ytp.LocaAddrs

		if err := ec.Encode(pi); err != nil {
			return err
		}
		return nil
	}
}
