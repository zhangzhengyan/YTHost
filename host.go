package host

import (
	"github.com/gogo/protobuf/proto"
	mnet "github.com/multiformats/go-multiaddr-net"
	"google.golang.org/grpc"
	"net"
)

type Conn struct {

}

type Host interface {
	Connect() *Conn
}

type host struct {
	cfg *Config
	listener mnet.Listener
	server *grpc.Server
	conns *ConnManager
}
type option interface {
}

func NewHost(options... option) (*host,error){

	hst := new(host)
	hst.cfg = NewConfig()
	ls,err := mnet.Listen(hst.cfg.ListenAddr)
	if err != nil {
		return nil ,err
	}
	hst.listener = ls
	hst.server = grpc.NewServer()

	if err:=hst.server.Serve(hst.listener.(net.Listener));err != nil {
		return nil,err
	}
	return hst,nil
}

func (hst *host)RegisterMsg(msgId int,pb proto.Message){}

func (hst *host)Serve(){
	for {
		if conn,err := hst.listener.Accept();err != nil {
			panic(err)
		} else {
			hst.conns.Add(conn)
		}
	}
}

func (hst *host)handle(conn net.Conn){
}
