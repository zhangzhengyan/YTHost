package host

import (
	"context"
	"fmt"
	"github.com/graydream/YTHost/client"
	"github.com/graydream/YTHost/clientStore"
	"github.com/graydream/YTHost/config"
	"github.com/graydream/YTHost/option"
	"github.com/graydream/YTHost/peerInfo"
	"github.com/graydream/YTHost/service"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	mnet "github.com/multiformats/go-multiaddr-net"

	"net/rpc"
	"sync"
)

//type Host interface {
//	Accept()
//	Addrs() []multiaddr.Multiaddr
//	Server() *rpc.Server
//	Config() *config.Config
//	Connect(ctx context.Context, pid peer.ID, mas []multiaddr.Multiaddr) (*client.YTHostClient, error)
//	RegisterHandler(id service.MsgId, handlerFunc service.Handler)
//}

type host struct {
	cfg      *config.Config
	listener mnet.Listener
	srv      *rpc.Server
	service.HandlerMap
	clientStore *clientStore.ClientStore
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

	hst.HandlerMap = make(service.HandlerMap)

	hst.clientStore = clientStore.NewClientStore(hst.Connect)
	return hst, nil
}

func (hst *host) Accept() {
	addrService := new(service.AddrService)
	addrService.Info.ID = hst.cfg.ID
	addrService.Info.Addrs = hst.Addrs()
	addrService.PubKey = hst.Config().Privkey.GetPublic()

	msgService := new(service.MsgService)
	msgService.Handler = hst.HandlerMap
	msgService.Pi = peerInfo.PeerInfo{hst.cfg.ID, hst.Addrs()}

	if err := hst.srv.RegisterName("as", addrService); err != nil {
		panic(err)
	}

	if err := hst.srv.RegisterName("ms", msgService); err != nil {
		panic(err)
	}

	hst.srv.Accept(mnet.NetListener(hst.listener))
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

func (hst *host) ClientStore() *clientStore.ClientStore {
	return hst.clientStore
}

func (hst *host) Addrs() []multiaddr.Multiaddr {

	port, err := hst.listener.Multiaddr().ValueForProtocol(multiaddr.P_TCP)
	if err != nil {
		return nil
	}

	tcpMa, err := multiaddr.NewMultiaddr(fmt.Sprintf("/tcp/%s", port))
	if err != nil {
		return nil
	}

	var res []multiaddr.Multiaddr
	maddrs, err := mnet.InterfaceMultiaddrs()
	if err != nil {
		return nil
	}

	for _, ma := range maddrs {
		newMa := ma.Encapsulate(tcpMa)
		if mnet.IsIPLoopback(newMa) {
			continue
		}
		res = append(res, newMa)
	}
	return res
}

// Connect 连接远程节点
func (hst *host) Connect(ctx context.Context, pid peer.ID, mas []multiaddr.Multiaddr) (*client.YTHostClient, error) {
	connChan := make(chan mnet.Conn)
	errorAll := make(chan struct{})

	wg := sync.WaitGroup{}
	wg.Add(len(mas))

	for _, addr := range mas {
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
		ytclt, err := client.WarpClient(clt, &peer.AddrInfo{
			hst.cfg.ID,
			hst.Addrs(),
		})
		if err != nil {
			return nil, err
		}
		return ytclt, nil
	case <-errorAll:
		return nil, fmt.Errorf("all maddr dail fail")
	case <-ctx.Done():
		return nil, fmt.Errorf("ctx time out")
	}
}

// ConnectAddrStrings 连接字符串地址
func (hst *host) ConnectAddrStrings(ctx context.Context, id string, addrs []string) (*client.YTHostClient, error) {
	pid, err := peer.IDFromString(id)
	if err != nil {
		return nil, err
	}

	var mas = make([]multiaddr.Multiaddr, len(addrs))
	for k, v := range addrs {
		ma, err := multiaddr.NewMultiaddr(v)
		if err != nil {
			continue
		}
		mas[k] = ma
	}

	return hst.Connect(ctx, pid, mas)
}

// SendMsg 发送消息
func (hst *host) SendMsg(ctx context.Context, pid peer.ID, mid int32, msg []byte) ([]byte, error) {
	clt, ok := hst.ClientStore().GetClient(pid)
	if !ok {
		return nil, fmt.Errorf("no client ID is:%s", pid.Pretty())
	}
	return clt.SendMsgClose(ctx, mid, msg)
}
