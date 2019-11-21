package host

import (
	"context"
	"fmt"
	"github.com/graydream/YTHost/YTHostError"
	"github.com/graydream/YTHost/config"
	"github.com/graydream/YTHost/event"
	"github.com/graydream/YTHost/option"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	mnet "github.com/multiformats/go-multiaddr-net"
	"sync"
)

type Conn struct {

}

type Host interface {
	Connect() *Conn
}

type host struct {
	cfg *config.Config
	listener mnet.Listener
	conns *ConnManager
	event.EventTrigger
}

func NewHost(options... option.Option) (*host,error){
	hst := new(host)
	hst.cfg = config.NewConfig()

	for _,bindOp:=range options{
		bindOp(hst.cfg)
	}
	ls,err := mnet.Listen(hst.cfg.ListenAddr)
	if err != nil {
		return nil,err
	}
	hst.listener = ls
	return hst,nil
}

func (hst *host)Listenner()mnet.Listener{
	return hst.listener
}

// Connect 连接远程节点
func (hst *host)Connect(ctx context.Context,mas []multiaddr.Multiaddr)(mnet.Conn,*YTHostError.YTError){
	var connChan = make(chan mnet.Conn)
	var AllDailFailErrorChan = make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(len(mas))
	for _,ma:=range mas{
		go func() {
			if conn,err := mnet.Dial(ma);err==nil{
				connChan <-conn
			} else {
				fmt.Println(err)
				wg.Done()
			}
		}()
	}
	go func() {
		wg.Wait()
		AllDailFailErrorChan<- struct{}{}
	}()
	select {
	case conn := <-connChan:
		return conn,nil
	case <-AllDailFailErrorChan:
		return nil,YTHostError.NewError(0,"All maddr dail fail")
	case <-ctx.Done():
		return nil,YTHostError.NewError(1,"dail timeout")
	}
}

func (hst *host)ConnectMaString(ctx context.Context,mastring string) (mnet.Conn,*YTHostError.YTError){
	if ma,err := multiaddr.NewMultiaddr(mastring);err != nil {
		return nil,YTHostError.NewError(0,fmt.Sprintf("Parse maddr Error %s",mastring))
	} else if pi,err:=peer.AddrInfoFromP2pAddr(ma);err!=nil {
		return nil,YTHostError.NewError(1,fmt.Sprintf("Parse maddr to peerInfo Error %s",mastring))
	} else {
		return hst.Connect(ctx,pi.Addrs)
	}
}

