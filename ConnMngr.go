package host

import (
	mnet "github.com/multiformats/go-multiaddr-net"
)

type ConnManager struct {
	Conns []mnet.Conn
}

func NewConnMngr() *ConnManager{
	connMngr:=new(ConnManager)
	return connMngr
}

func (cm *ConnManager)Add(conn mnet.Conn){
	cm.Conns = append(cm.Conns,conn)
}
