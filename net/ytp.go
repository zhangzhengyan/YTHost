package net

import (
	"bufio"
	"context"
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
// \n\n 是终止标志
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
	sc := bufio.NewScanner(ytp)

	if err := ytp.SendPeerInfo(ctx); err != nil {
		return nil
	}

	for sc.Scan() {
		select {
		case <-ctx.Done():
			break
		default:
			switch sc.Text() {
			//case "Is YTHost?":
			//	if _,err:=w.WriteString(fmt.Sprintln("YTHost: 0.0.1"));err!=nil{
			//		return err
			//	}
			//case "YTHost: 0.0.1":
			//	ytp.RemoteIsYTHost = true
			//	if err:=ytp.sendPeerInfo(ctx);err!=nil {
			//		return err
			//	}
			case "\n\n":
				ytp.Finish = true
				break
			default:
				line := sc.Text()
				var id string
				if n, err := fmt.Sscanf(line, "ID:%s\n", &id); err == nil && n > 0 {
					fmt.Println(id)
					if pid, err := peer.IDB58Decode(id); err != nil {

						return err
					} else {
						ytp.RemoteID = pid
					}
				}
				var addr string
				if n, err := fmt.Sscanf(line, "Addr:%s\n", &addr); err == nil && n > 0 {
					fmt.Println(addr)
					if ma, err := multiaddr.NewMultiaddr(addr); err == nil {
						ytp.RemoteAddrs = append(ytp.RemoteAddrs, ma)
					}
				}
			}
		}
	}
	return nil
}

func (ytp *YTP) SendPeerInfo(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("ctx time out")
	default:
		w := bufio.NewWriter(ytp)
		if _, err := w.WriteString(fmt.Sprintf("ID:%s\n", ytp.LocalID.Pretty())); err != nil {
			return err
		}
		for _, addr := range ytp.LocaAddrs {
			if _, err := w.WriteString(fmt.Sprintf("Addr:%s\n", addr.String())); err != nil {
				return err
			}
		}
		// 结束
		if _, err := w.WriteString(fmt.Sprintf("\n\n")); err != nil {
			return err
		}
	}
	return nil
}
