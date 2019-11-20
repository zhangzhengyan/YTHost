package host

import (
	net "github.com/multiformats/go-multiaddr-net"
)

type ConnManager struct {
	Conns []net.Conn
}

func NewConnMngr() *ConnManager{
	connMngr:=new(ConnManager)
	return connMngr
}

func (cm *ConnManager)Add(conn net.Conn){
	cm.Conns = append(cm.Conns,conn)
}
