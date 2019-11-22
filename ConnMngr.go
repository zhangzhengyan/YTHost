package host

import (
	"context"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/graydream/YTHost/event"
	"github.com/graydream/YTHost/pbMsgHandler"
	mnet "github.com/multiformats/go-multiaddr-net"
	"time"
)

type ConnManager struct {
	conns map[string]*pbMsgHandler.PBMsgHandler
	pbMsgHandler.PBMsgHanderMap
	event.EventTrigger
}

func NewConnMngr() *ConnManager{
	connMngr:=new(ConnManager)
	connMngr.conns = make(map[string]*pbMsgHandler.PBMsgHandler)
	return connMngr
}

func (cm *ConnManager)addConn(pid string,conn mnet.Conn){
	cm.conns[pid]=pbMsgHandler.NewPBMsgHander(conn)
}

func (cm *ConnManager)Conns() map[string]*pbMsgHandler.PBMsgHandler{
	return cm.conns
}

func (cm *ConnManager)serve(ctx context.Context){
	for {
		if len(cm.conns) == 0 {
			time.Sleep(500*time.Millisecond)
		} else {
			for _,v:=range cm.conns{
				if msg,err:=v.Accept();err == nil {
					if err := cm.Call(msg.ID,msg.Data,v);err != nil {
						cm.Emit(event.Event{"error",err})
					}
				} else {
					cm.Emit(event.Event{"error",err})
				}
			}
		}
	}
}

func (cm *ConnManager)SendMsg(pid string, msgId uint16, data proto.Message) error {
	if conn,ok:=cm.conns[pid];ok{
		return conn.SendMsg(msgId,data)
	} else {
		return fmt.Errorf("No pid connect")
	}
	return nil
}

func (cm *ConnManager)DisConnect(pid string)error{
	if conn,ok:=cm.conns[pid];ok {
		delete(cm.conns,pid)
		err:= conn.Close()
		return err
	}
	return fmt.Errorf("No connnect:%s",pid)
}
