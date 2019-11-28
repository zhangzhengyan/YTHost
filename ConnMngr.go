package host

import (
	"context"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/graydream/YTHost/event"
	"github.com/graydream/YTHost/net"
	"github.com/graydream/YTHost/pbMsgHandler"
	"github.com/libp2p/go-libp2p-core/peer"
	"time"
)

type ConnManager struct {
	conns map[peer.ID]*pbMsgHandler.PBMsgHandler
	pbMsgHandler.PBMsgHanderMap
	event.EventTrigger
}

func NewConnMngr() *ConnManager {
	connMngr := new(ConnManager)
	connMngr.conns = make(map[peer.ID]*pbMsgHandler.PBMsgHandler)
	return connMngr
}

func (cm *ConnManager) addConn(pi peer.AddrInfo, conn net.Conn) error {
	cm.conns[conn.RemotePeer().ID] = pbMsgHandler.NewPBMsgHander(conn)
	fmt.Println("peerID", conn.RemotePeer().ID)
	return nil
}

func (cm *ConnManager) Conns() map[peer.ID]*pbMsgHandler.PBMsgHandler {
	return cm.conns
}

func (cm *ConnManager) serve(ctx context.Context) {
	for {
		if len(cm.conns) == 0 {
			time.Sleep(500 * time.Millisecond)
		} else {
			for _, v := range cm.conns {
				if msg, err := v.Accept(); err == nil {
					if err := cm.Call(msg.ID, msg.Data, v); err != nil {
						cm.Emit(event.Event{"error", err})
					}
				} else {
					cm.Emit(event.Event{"error", err})
				}
			}
		}
	}
}

func (cm *ConnManager) SendMsg(pid peer.ID, msgId uint16, data proto.Message) error {
	if conn, ok := cm.conns[pid]; ok {
		return conn.SendMsg(msgId, data)
	} else {
		return fmt.Errorf("No pid connect")
	}
	return nil
}

func (cm *ConnManager) DisConnect(pid peer.ID) error {
	if conn, ok := cm.conns[pid]; ok {
		delete(cm.conns, pid)
		err := conn.Close()
		return err
	}
	return fmt.Errorf("No connnect:%s", pid)
}
