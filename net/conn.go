package net

import (
	"github.com/multiformats/go-multiaddr-net"
)

type Conn interface {
	manet.Conn
	ID () string
}

type ytconn struct {
	manet.Conn
	id string
}

func (yc *ytconn)ID() string {
	return yc.id
}

//func WarpConn(conn manet.Conn) Conn{
//	yc := ytconn{
//		conn,
//	}
//	return yc
//}
