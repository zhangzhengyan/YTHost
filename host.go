package host

import (
	"context"
	"fmt"
	"github.com/yottachain/YTHost/client"
	ci "github.com/yottachain/YTHost/clientInterface"
	"github.com/yottachain/YTHost/clientStore"
	"github.com/yottachain/YTHost/config"
	"github.com/yottachain/YTHost/ioStream"
	"github.com/yottachain/YTHost/option"
	"github.com/yottachain/YTHost/peerInfo"
	"github.com/yottachain/YTHost/service"
	"github.com/libp2p/go-libp2p-core/peer"
	reuse "github.com/libp2p/go-reuseport"
	"github.com/mr-tron/base58"
	"github.com/multiformats/go-multiaddr"
	mnet "github.com/multiformats/go-multiaddr-net"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"net/rpc"
	//"github.com/yottachain/YTHost/rpc"
	"sync"
	"time"
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
	listener net.Listener
	srv      *rpc.Server
	service.HandlerMap
	clientStore *clientStore.ClientStore
}

func Listen(laddr multiaddr.Multiaddr) (net.Listener, error) {
	// get the net.Listen friendly arguments from the remote addr
	lnet, lnaddr, err := mnet.DialArgs(laddr)
	if err != nil {
		return nil, err
	}

	//nl, err := net.Listen(lnet, lnaddr)
	nl, err := reuse.Listen(lnet, lnaddr)
	if err != nil {
		return nil, err
	}

	// we want to fetch the new multiaddr from the listener, as it may
	// have resolved to some other value. WrapNetListener does it for us.
	return nl, nil
}

func NewHost(options ...option.Option) (*host, error) {
	hst := new(host)
	hst.cfg = config.NewConfig()

	for _, bindOp := range options {
		bindOp(hst.cfg)
	}

	//ls, err := mnet.Listen(hst.cfg.ListenAddr)
	ls, err := Listen(hst.cfg.ListenAddr)

	if err != nil {
		return nil, err
	}

	hst.listener = ls

	srv := rpc.NewServer()
	hst.srv = srv

	hst.HandlerMap = make(service.HandlerMap)

	hst.clientStore = clientStore.NewClientStore(hst.Connect)

	if hst.cfg.PProf != "" {
		go func() {
			if err := http.ListenAndServe(hst.cfg.PProf, nil); err != nil {
				fmt.Println("PProf open fail:", err)
			} else {
				fmt.Println("PProf debug open:", hst.cfg.PProf)
			}
		}()
	}

	return hst, nil
}

func (hst *host) Accept() {
	addrService := new(service.AddrService)
	addrService.Info.ID = hst.cfg.ID
	addrService.Info.Addrs = hst.Addrs()
	addrService.PubKey = hst.Config().Privkey.GetPublic()

	msgService := new(service.MsgService)
	msgService.Handler = hst.HandlerMap
	msgService.LocalPriKey = hst.cfg.Privkey
	msgService.Pi = peerInfo.PeerInfo{hst.cfg.ID, hst.Addrs()}
	msgService.ClientStore = hst.clientStore
	msgService.LocalPeerID = hst.cfg.ID

	if err := hst.srv.RegisterName("as", addrService); err != nil {
		panic(err)
	}

	if err := hst.srv.RegisterName("ms", msgService); err != nil {
		panic(err)
	}

	//hst.srv.Accept(mnet.NetListener(hst.listener))
	/*
	这里不调用rpc.server 的accept
	因为我们要根据accept到的连接建立新的rpc client
	并维护这个client表，用于中继转发
	*/
	//lis := mnet.NetListener(hst.listener)
	lis := hst.listener
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Print("rpc.Serve: accept:", err.Error())
			return
		}

		//go hst.srv.ServeConn(conn)
		sConn, cConn := ioStream.NewStreamHandler(conn)
		go hst.srv.ServeConn(sConn)

		go func(sConn *ioStream.ReadWriteCloser, cConn *ioStream.ReadWriteCloser) {
			//connect := conn
			clt := rpc.NewClient(cConn)
			if nil == clt {
				return
			}
			tryCount := 1
			var ytclt ci.YTHClient
			for {
				if tryCount > 3 {
					break
				}else {
					tryCount++
				}
				ytclt, err = client.WarpClient(clt, &peer.AddrInfo{
					hst.cfg.ID,
					hst.Addrs(),
				}, hst.cfg.Privkey.GetPublic(), peer.ID(0))			//这种情况目标ID就是remoteID

				if nil != err {
					fmt.Println("rpc.Serve: accept conn warpClient:", err.Error())
					continue
				}

				hst.clientStore.Store(ytclt.GetRemotePeerID(), ytclt)
				break
			}

			tryCount = 1
			for {
				if tryCount > 6 {
					_ = sConn.Close()
					_ = cConn.Close()
					//_ = hst.clientStore.Close(ytclt.RemotePeerID)
					//_ = connect.Close()
					break
				}else {
					tryCount++
				}
				if ytclt != nil {
						ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
						defer cancel()
						if ytclt.Ping(ctx) {
							tryCount--
							time.Sleep(time.Second*10)
						}
				}
			}
		}(sConn, cConn)
	}
}

func (hst *host) Listenner() net.Listener {
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
	lis, err := mnet.WrapNetListener(hst.listener)
	if err != nil {
		return nil
	}
	//port, err := hst.listener.Multiaddr().ValueForProtocol(multiaddr.P_TCP)
	port, err := lis.Multiaddr().ValueForProtocol(multiaddr.P_TCP)
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
func (hst *host) Connect(ctx context.Context, pid peer.ID, mas []multiaddr.Multiaddr) (ci.YTHClient, error) {
	conn, err := hst.connect(ctx, pid, mas)
	if err != nil {
		return nil, err
	}

	sConn, cConn := ioStream.NewStreamHandler(conn)
	go hst.srv.ServeConn(sConn)
	clt := rpc.NewClient(cConn)

	ytclt, err := client.WarpClient(clt, &peer.AddrInfo{
		hst.cfg.ID,
		hst.Addrs(),
	}, hst.cfg.Privkey.GetPublic(), pid)

	if nil != err {
		return nil, err
	}

	go func() {
		tryCount := 1
		for {
			if tryCount > 6 {
				_ = sConn.Close()
				_ = cConn.Close()
				//_ = hst.clientStore.Close(ytclt.RemotePeerID)
				//_ = connect.Close()
				break
			}else {
				tryCount++
			}
			if ytclt != nil {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
				defer cancel()
				if ytclt.Ping(ctx) {
					tryCount--
					time.Sleep(time.Second*10)
				}
			}
		}
	}()

	return ytclt, nil
}

func (hst *host) connect(ctx context.Context, pid peer.ID, mas []multiaddr.Multiaddr) (net.Conn, error) {
	connChan := make(chan net.Conn)
	errChan := make(chan error)
	wg := sync.WaitGroup{}
	wg.Add(len(mas))

	lis, err := mnet.WrapNetListener(hst.listener)
	if err != nil {
		return nil, err
	}
	//port, err := hst.listener.Multiaddr().ValueForProtocol(multiaddr.P_TCP)
	port, err := lis.Multiaddr().ValueForProtocol(multiaddr.P_TCP)
	if err != nil {
		return nil, err
	}
	lnsaddr := fmt.Sprintf(":%s", port)

	go func() {
		wg.Wait()
		select {
		case errChan <- fmt.Errorf("dail all maddr fail"):
		case <-time.After(time.Millisecond * 500):
			return
		}
	}()

	for _, addr := range mas {
		go func(addr multiaddr.Multiaddr) {
			defer wg.Done()
			lnet, lndaddr, err := mnet.DialArgs(addr)
			if err != nil {
				if hst.cfg.Debug {
					log.Println("conn error:", err)
				}
			}
			//if conn, err := (&mnet.Dialer{}).DialContext(ctx, addr); err == nil {
			if conn, err := reuse.Dial(lnet, lnsaddr, lndaddr); err == nil {
				select {
				case connChan <- conn:
				case <-time.After(time.Second * 5):
				}
			} else {
				if hst.cfg.Debug {
					log.Println("conn error:", err)
				}
			}
		}(addr)
	}

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("ctx quit")
		case conn := <-connChan:
			return conn, nil
		case err := <-errChan:
			return nil, err
		}
	}
}

// ConnectAddrStrings 连接字符串地址
func (hst *host) ConnectAddrStrings(ctx context.Context, id string, addrs []string) (ci.YTHClient, error) {

	buf, _ := base58.Decode(id)
	pid, err := peer.IDFromBytes(buf)
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
	return clt.SendMsg(ctx, mid, msg)
}
