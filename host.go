package host

import (
	"context"
	"fmt"
	"github.com/graydream/YTHost/config"
	"github.com/graydream/YTHost/option"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	mnet "github.com/multiformats/go-multiaddr-net"

	"net/rpc"
	"sync"
)

//type Host interface {
//	Connect() net.Conn
//}

type host struct {
	cfg      *config.Config
	listener mnet.Listener
	srv      *rpc.Server
}

func NewHost(options ...option.Option) (*host, error) {
	hst := new(host)
	hst.cfg = config.NewConfig()

	for _, bindOp := range options {
		bindOp(hst.cfg)
	}

	ls, err := mnet.Listen(hst.cfg.ListenAddr)

	if err != nil {
		return nil, err
	}

	hst.listener = ls

	srv := rpc.NewServer()
	hst.srv = srv

	return hst, nil
}

func (hst *host) Accept() {
	mnet.NetListener(hst.listener).Accept()
}

func (hst *host) Listenner() mnet.Listener {
	return hst.listener
}
func (hst *host) Server() *rpc.Server {
	return hst.srv
}

func (hst *host) Config() *config.Config {
	return hst.cfg
}

func (hst *host) Addrs() ([]multiaddr.Multiaddr, error) {

	port, err := hst.listener.Multiaddr().ValueForProtocol(multiaddr.P_TCP)
	if err != nil {
		return nil, err
	}

	tcpMa, err := multiaddr.NewMultiaddr(fmt.Sprintf("/tcp/%s", port))
	if err != nil {
		return nil, err
	}

	var res []multiaddr.Multiaddr
	maddrs, err := mnet.InterfaceMultiaddrs()
	if err != nil {
		return nil, err
	}

	for _, ma := range maddrs {
		newMa := ma.Encapsulate(tcpMa)
		if mnet.IsIPLoopback(newMa) {
			continue
		}
		res = append(res, newMa)
	}
	return res, nil
}

// Connect 连接远程节点
func (hst *host) Connect(ctx context.Context, pid peer.ID, mas []multiaddr.Multiaddr) (*rpc.Client, error) {
	connChan := make(chan mnet.Conn)
	errorAll := make(chan struct{})

	wg := sync.WaitGroup{}
	addrs, err := hst.Addrs()
	if err != nil {
		return nil, err
	}
	wg.Add(len(addrs))

	for _, addr := range addrs {
		go func(ma multiaddr.Multiaddr) {
			if conn, err := mnet.Dial(ma); err != nil {
				wg.Done()
			} else {
				connChan <- conn
			}
		}(addr)
	}
	go func() {
		wg.Wait()
		errorAll <- struct{}{}
	}()

	select {
	case conn := <-connChan:
		clt := rpc.NewClient(conn)
		return clt, nil
	case <-errorAll:
		return nil, fmt.Errorf("all maddr dail fail")
	case <-ctx.Done():
		return nil, fmt.Errorf("ctx time out")
	}
}
