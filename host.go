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
				if conn, err := hst.listener.Accept(); err != nil {
					hst.Emit(event.Event{"error", err})
				} else {
					maddrs, _ := hst.Addrs()
					pi := peer.AddrInfo{ID: hst.cfg.ID, Addrs: maddrs}
					hst.addConn(pi, conn)
				}
			}
		}
	}()
	hst.serve(ctx)
}

// Connect 连接远程节点
func (hst *host) Connect(ctx context.Context, pid peer.ID, mas []multiaddr.Multiaddr) (mnet.Conn, *YTHostError.YTError) {
	var connChan = make(chan mnet.Conn)
	var AllDailFailErrorChan = make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(len(mas))
	for _, ma := range mas {
		go func() {
			if conn, err := mnet.Dial(ma); err == nil {
				connChan <- conn
			} else {
				fmt.Println(err)
				wg.Done()
			}
		}()
	}
	go func() {
		wg.Wait()
		AllDailFailErrorChan <- struct{}{}
	}()
	select {
	case conn := <-connChan:
		maddrs, _ := hst.Addrs()
		pi := peer.AddrInfo{ID: hst.cfg.ID, Addrs: maddrs}
		if err := hst.addConn(pi, conn); err != nil {
			return nil, YTHostError.NewError(2, "add conn fail")
		}
		return conn, nil
	case <-AllDailFailErrorChan:
		return nil, YTHostError.NewError(0, "All maddr dail fail")
	case <-ctx.Done():
		return nil, YTHostError.NewError(1, "dail timeout")
	}
}

func (hst *host) ConnectMaString(ctx context.Context, pid peer.ID, mastring string) (mnet.Conn, *YTHostError.YTError) {
	if ma, err := multiaddr.NewMultiaddr(mastring); err != nil {
		return nil, YTHostError.NewError(0, fmt.Sprintf("Parse maddr Error %s", mastring))
	} else if pi, err := peer.AddrInfoFromP2pAddr(ma); err != nil {
		return nil, YTHostError.NewError(1, fmt.Sprintf("Parse maddr to peerInfo Error %s", mastring))
	} else {
		return hst.Connect(ctx, pid, pi.Addrs)
	}
}
