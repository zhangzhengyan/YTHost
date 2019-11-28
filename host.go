package host

import (
	"context"
	"fmt"
	"github.com/graydream/YTHost/config"
	"github.com/graydream/YTHost/option"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	mnet "github.com/multiformats/go-multiaddr-net"
	"sync"
)

//type Host interface {
//	Connect() net.Conn
//}

type host struct {
	cfg      *config.Config
	listener mnet.Listener
	*ConnManager
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
	hst.ConnManager = NewConnMngr()
	return hst, nil
}

func (hst *host) Listenner() mnet.Listener {
	return hst.listener
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

func (hst *host) Serve(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()
	hst.serve(ctx)
}

// Connect 连接远程节点
func (hst *host) Connect(ctx context.Context, pid peer.ID, mas []multiaddr.Multiaddr) (mnet.Conn, error) {
	var connChan = make(chan mnet.Conn)
	var AllDailFailErrorChan = make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(len(mas))

	go func() {
		for _, ma := range mas {
			go func(ma multiaddr.Multiaddr) {
				if conn, err := mnet.Dial(ma); err == nil {
					connChan <- conn
				} else {
					fmt.Println(err)
					wg.Done()
				}
			}(ma)
		}
	}()

	go func() {
		wg.Wait()
		AllDailFailErrorChan <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("dail timeout")
	case conn := <-connChan:
		//maddrs, _ := hst.Addrs()
		//pi := peer.AddrInfo{ID: hst.cfg.ID, Addrs: maddrs}
		return conn, nil
	case <-AllDailFailErrorChan:
		return nil, fmt.Errorf("All maddr dail fail")
	}
}

func (hst *host) ConnectMaString(ctx context.Context, pid peer.ID, mastring string) (mnet.Conn, error) {
	if ma, err := multiaddr.NewMultiaddr(mastring); err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Parse maddr Error %s", mastring))
	} else if pi, err := peer.AddrInfoFromP2pAddr(ma); err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Parse maddr to peerInfo Error %s", mastring))
	} else {
		return hst.Connect(ctx, pid, pi.Addrs)
	}
}
