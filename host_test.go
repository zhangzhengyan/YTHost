package host_test

import (
	"context"
	ra "crypto/rand"
	"fmt"
	host "github.com/graydream/YTHost"
	"github.com/graydream/YTHost/client"
	. "github.com/graydream/YTHost/hostInterface"
	"github.com/graydream/YTHost/option"
	"github.com/graydream/YTHost/service"
	ic "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/multiformats/go-multiaddr"
	"math/rand"
	"sync"
	"testing"
	"time"
)

// 测试创建通讯节点
func TestNewHost(t *testing.T) {
	var localMa2 = "/ip4/0.0.0.0/tcp/9003"

	pk , _, _ := ic.GenerateSecp256k1Key(ra.Reader)		//option Identity cover
	ma, _ := multiaddr.NewMultiaddr(localMa2)			// option listenaddr cover
	if hst, err := host.NewHost(option.ListenAddr(ma), option.Identity(pk)); err != nil {
		t.Fatal(err)
	} else {
		maddrs := hst.Addrs()
		for _, ma := range maddrs {
			t.Log(ma)
		}
		t.Log(hst.Config().ID)
	}

}

type RpcService struct {
}

type Reply struct {
	Value string
}

func (rs *RpcService) Test(req string, res *Reply) error {
	res.Value = "pong"
	return nil
}

// GetRandomLocalMutlAddr 获取随机本地地址
func GetRandomLocalMutlAddr() multiaddr.Multiaddr {
	port := rand.Int()%1000 + 9000
	mastr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)
	ma, _ := multiaddr.NewMultiaddr(mastr)
	return ma
}

func GetRandomHost() Host {
	return GetHost(GetRandomLocalMutlAddr())
}
func GetHost(ma multiaddr.Multiaddr) Host {
	hst, err := host.NewHost(option.ListenAddr(ma))
	if err != nil {
		panic(err)
	}
	return hst
}

// 测试建立连接发送消息
func TestConn(t *testing.T) {

	hst := GetRandomHost()

	// 注册远程接口
	if err := hst.Server().Register(new(RpcService)); err != nil {
		t.Fatal(err)
	}

	go hst.Accept()

	hst2 := GetRandomHost()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	clt, err := hst2.Connect(ctx, hst.Config().ID, hst.Addrs())
	if err != nil {
		t.Fatal(err.Error())
	}

	var res Reply

	// 调用远程接口
	if err := clt.Call("RpcService.Test", "", &res); err != nil {
		t.Fatal(err)
	} else {
		t.Log(res.Value)
	}

}

// 测试建立连接，交换peerinfo
func TestConnSendPeerInfo(t *testing.T) {
	hst := GetRandomHost()
	if err := hst.Server().Register(new(RpcService)); err != nil {
		t.Fatal(err)
	}

	go hst.Accept()

	hst2 := GetRandomHost()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	clt, err := hst2.Connect(ctx, hst.Config().ID, hst.Addrs())
	if err != nil {
		t.Fatal(err.Error())
	}

	peerInfo := clt.RemotePeer()

	// 打印节点信息
	t.Log(peerInfo.ID.Pretty(), peerInfo.Addrs, clt.RemotePeerPubkey())
}

// 发送，处理消息
func TestHandleMsg(t *testing.T) {
	hst := GetRandomHost()
	if err := hst.Server().Register(new(RpcService)); err != nil {
		t.Fatal(err)
	}

	hst.RegisterGlobalMsgHandler(func(requestData []byte, head service.Head) (bytes []byte, e error) {
		t.Log(string(requestData), head.RemotePubKey, head.RemotePeerID, head.RemoteAddrs, fmt.Sprintf("0x%0x", head.MsgId))
		return []byte("111"), nil
	})

	go hst.Accept()

	hst2 := GetRandomHost()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	clt, err := hst2.Connect(ctx, hst.Config().ID, hst.Addrs())
	// 关闭连接
	defer clt.Close()
	if err != nil {
		t.Fatal(err.Error())
	}
	if res, err := clt.SendMsg(context.Background(), 0x11, []byte("2222")); err != nil {
		t.Fatal(err)
	} else {
		t.Log(string(res))
	}
}

// 发送，处理全局消息
func TestGlobalHandleMsg(t *testing.T) {
	hst := GetRandomHost()
	if err := hst.Server().Register(new(RpcService)); err != nil {
		t.Fatal(err)
	}

	hst.RegisterGlobalMsgHandler(func(requestData []byte, head service.Head) (bytes []byte, e error) {
		t.Log(string(requestData), head.RemotePubKey, head.RemotePeerID, head.RemoteAddrs)
		return []byte("111"), nil
	})

	go hst.Accept()

	hst2 := GetRandomHost()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	clt, err := hst2.Connect(ctx, hst.Config().ID, hst.Addrs())
	// 关闭连接
	defer clt.Close()
	if err != nil {
		t.Fatal(err.Error())
	}
	if res, err := clt.SendMsg(context.Background(), 0x11, []byte("2222")); err != nil {
		t.Fatal(err)
	} else {
		t.Log(string(res))
	}
	if res, err := clt.SendMsg(context.Background(), 0x12, []byte("121212")); err != nil {
		t.Fatal(err)
	} else {
		t.Log(res)
	}
}

// 发送，处理全局消息,自动关闭连接
func TestGlobalHandleMsgClose(t *testing.T) {
	hst := GetRandomHost()
	if err := hst.Server().Register(new(RpcService)); err != nil {
		t.Fatal(err)
	}

	hst.RegisterGlobalMsgHandler(func(requestData []byte, head service.Head) (bytes []byte, e error) {
		t.Log(string(requestData), head.RemotePubKey, head.RemotePeerID, head.RemoteAddrs)
		return []byte("111"), nil
	})

	go hst.Accept()

	hst2 := GetRandomHost()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	clt, err := hst2.Connect(ctx, hst.Config().ID, hst.Addrs())
	if err != nil {
		t.Fatal(err.Error())
	}
	if clt.Ping(ctx) {
		t.Log("success")
	} else {
		t.Fatal("ping fail")
	}
	if res, err := clt.SendMsgClose(context.Background(), 0x11, []byte("2222")); err != nil {
		t.Fatal(err)
	} else {
		t.Log(string(res))
	}
	if res, err := clt.SendMsg(context.Background(), 0x12, []byte("121212")); err == nil {
		t.Fatal(fmt.Errorf("此处连接应该关闭"), res)
	} else {
		t.Log(err)
	}

	//通过字符串地址连接一次
	maddrs := hst.Addrs()
	addrs := make([]string, len(maddrs))
	for k, m := range maddrs {
		addrs[k] = m.String()
	}

	clt, err = hst2.ConnectAddrStrings(ctx, hst.Config().ID.String(), addrs)
	if err != nil {
		t.Fatal(err.Error())
	}
	if res, err := clt.SendMsgClose(context.Background(), 0x11, []byte("2222")); err != nil {
		t.Fatal(err)
	} else {
		t.Log(string(res))
	}

}

// 测试连接管理
func TestCS(t *testing.T) {
	hst := GetRandomHost()

	if err := hst.Server().Register(new(RpcService)); err != nil {
		t.Fatal(err)
	}

	hst.RegisterGlobalMsgHandler(func(requestData []byte, head service.Head) (bytes []byte, e error) {
		t.Log(string(requestData), head.RemotePubKey, head.RemotePeerID, head.RemoteAddrs)
		return []byte("111"), nil
	})

	go hst.Accept()

	hst2 := GetRandomHost()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	clt, err := hst2.ClientStore().Get(ctx, hst.Config().ID, hst.Addrs())
	if err != nil {
		t.Fatal(err)
	}

	// 第二次获取相同连接，不应建立新的连接
	_, err = hst2.ClientStore().Get(ctx, hst.Config().ID, hst.Addrs())
	if err != nil {
		t.Fatal(err)
	}
	if hst2.ClientStore().Len() > 1 {
		t.Fatal("不应该建立新连接")
	}

	// 另一种发送消息的方式
	if res, err := hst2.SendMsg(context.Background(), hst.Config().ID, 0x11, []byte("2222")); err != nil {
		t.Fatal(err)
	} else {
		t.Log(string(res))
	}
	if err := hst2.ClientStore().Close(hst.Config().ID); err != nil {
		t.Fatal(err)
	}
	if res, err := clt.SendMsg(context.Background(), 0x12, []byte("121213")); err == nil {
		t.Fatal(fmt.Errorf("此处连接应该关闭"), res)
	} else {
		t.Log(err)
	}
}


//测试多连接
func TestMutlConn(t *testing.T) {
	mastr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", 9000)
	ma, _ := multiaddr.NewMultiaddr(mastr)
	hst, err := host.NewHost(option.ListenAddr(ma))
	if err != nil {
		panic(err)
	}

	var count int64 = 0

	hst.RegisterGlobalMsgHandler(func(requestData []byte, head service.Head) (bytes []byte, e error) {
		count =  count+1
		//fmt.Println(fmt.Sprintf("msg num is %d", count), string(requestData))
		fmt.Println(fmt.Sprintf("msg num is %d", count))
		return []byte("111111111111"), nil
	})

	go hst.Accept()

	lenth := 5000
	hstcSilce := make([]Host, lenth)
	connSilce := make([]*client.YTHostClient, lenth)
	//wg := sync.WaitGroup{}
	//wg.Add(lenth)

	for i := 0; i < lenth; i++ {
		j := i
		port := 19000 + j + 1
		mastr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)
		ma, _ := multiaddr.NewMultiaddr(mastr)
		hstcSilce[j], err = host.NewHost(option.ListenAddr(ma))
		if err != nil {
			//panic(err)
			//maybe port by used
			hstcSilce[j] = nil
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		connSilce[j], err = hstcSilce[j].Connect(ctx, hst.Config().ID, hst.Addrs())
		if err != nil {
			t.Log(err.Error())
			continue
		}else{
			fmt.Println(connSilce[j].LocalPeer())
		}

		ctx1, cancel1 := context.WithCancel(context.Background())
		defer cancel1()
		go func(ctx context.Context, conn **client.YTHostClient, index int) (error) {
			for {
				select {
				case <-ctx.Done():
					return nil
				default:
						ctx_local, cancel_local := context.WithTimeout(context.Background(), time.Second*10)
						defer cancel_local()
						_, err := (*conn).SendMsg(ctx_local, 0x0, []byte("111111111"))
						if err != nil {
							fmt.Println(err)
						} else {
							fmt.Printf("conn[%d] send succeed\n", index)
						}
						time.Sleep(time.Second * 2)
				}
			}
		}(ctx1, &connSilce[j], j)

		//go func(ctx context.Context, conn *client.YTHostClient, index int) (error) {
		//	for {
		//		select {
		//		case <-ctx.Done():
		//			//wg.Done()
		//			return nil
		//		default:
		//				ctx_local, cancel_local := context.WithTimeout(context.Background(), time.Second*10)
		//				defer cancel_local()
		//				_, err := conn.SendMsg(ctx_local, 0x0, []byte("111111111"))
		//				if err != nil {
		//					fmt.Println(err)
		//				} else {
		//					fmt.Printf("conn[%d] send succeed\n", index)
		//					//time.Sleep(time.Second * 2)
		//				}
		//		}
		//	}
		//}(ctx1, connSilce[j], j)

	}

	//for {
	//	time.Sleep(time.Second * 10)
	//	index := rand.Int()%lenth
	//	for i := index; i < index + lenth/100; i++ {
	//		if connSilce[index] != nil && hstcSilce[index] != nil  {
	//			connSilce[index].Close()
	//			fmt.Printf("conn[%d] by closed\n", index)
	//		}else {
	//			continue
	//		}
	//	}
	//
	//
	//	time.Sleep(time.Second * 10)
	//	for i := index; i < index + lenth/100; i++ {
	//		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	//		defer cancel()
	//		connSilce[index], err = hstcSilce[index].Connect(ctx, hst.Config().ID, hst.Addrs())
	//		if err != nil {
	//			t.Fatal(err.Error())
	//		} else {
	//			fmt.Printf("conn[%d] by opened", index)
	//			fmt.Println(connSilce[index].LocalPeer())
	//		}
	//	}
	//}

	//wg.Wait()
	for {
		time.Sleep(time.Second*1)
	}
}

func TestMutlConn111(t *testing.T) {
	mastr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", 9000)
	ma, _ := multiaddr.NewMultiaddr(mastr)
	hst, err := host.NewHost(option.ListenAddr(ma))
	if err != nil {
		panic(err)
	}

	var count uint64 = 0

	hst.RegisterGlobalMsgHandler(func(requestData []byte, head service.Head) (bytes []byte, e error) {
		count =  count+1
		//fmt.Println(fmt.Sprintf("msg num is %d", count), string(requestData))
		fmt.Println(fmt.Sprintf("msg num is %d", count))
		return []byte("111111111111"), nil
	})

	go hst.Accept()

	lenth := 5000
	//hstcSilce := make([]Host, lenth)
	connSilce := make([]*client.YTHostClient, lenth)

	for k,_:=range connSilce{
		ma,_:=multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d",13000+k))
		h,err:=host.NewHost(option.ListenAddr(ma))
		if err != nil {
			fmt.Println(err)
		} else {
			conn,_:=h.Connect(context.Background(),hst.Config().ID,hst.Addrs())
			connSilce[k]=conn
		}
		time.After(time.Millisecond*100)
	}

	for {
		wg:=sync.WaitGroup{}
		wg.Add(lenth)
		for _,c:=range connSilce {
			go func() {
				if c==nil{
					return
				}
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				c.SendMsg(ctx, 0x0, []byte("msg"))
				defer wg.Done()
			}()
		}
		wg.Wait()
	}

	select {}
}

//测试多并发连接
func TestMutlconcurrentConn(t *testing.T) {
	mastr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", 9000)
	ma, _ := multiaddr.NewMultiaddr(mastr)
	hst, err := host.NewHost(option.ListenAddr(ma))
	if err != nil {
		panic(err)
	}

	var count int64 = 0

	hst.RegisterGlobalMsgHandler(func(requestData []byte, head service.Head) (bytes []byte, e error) {
		count =  count+1
		//fmt.Println(fmt.Sprintf("msg num is %d", count), string(requestData))
		fmt.Println(fmt.Sprintf("msg num is %d", count))
		return []byte("111111111111"), nil
	})

	go hst.Accept()

	lenth := 5000
	hstcSilce := make([]Host, lenth)
	connSilce := make([]*client.YTHostClient, lenth)

	for i := 0; i < lenth; i++ {
		j := i
		port := 19000 + j + 1
		mastr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)
		ma, _ := multiaddr.NewMultiaddr(mastr)
		hstcSilce[j], err = host.NewHost(option.ListenAddr(ma))
		if err != nil {
			//panic(err)
			//maybe port by used
			hstcSilce[j] = nil
			continue
		}
	}

	//test concurrent connect
	concurNum := 66
	if concurNum > lenth {
		concurNum = lenth
	}
	for i := 0; i < lenth; i = i + concurNum {
		curEndPos := i + concurNum
		stepSize := 0
		wg := sync.WaitGroup{}
		if curEndPos < lenth {
			stepSize = concurNum
		}else {
			stepSize = lenth - i
		}
		wg.Add(stepSize)

		for j := i; j < i + stepSize; j++ {
			idx := j
			go func(conn **client.YTHostClient, host *Host){
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
				defer cancel()
				*conn, err = (*host).ClientStore().Get(ctx, hst.Config().ID, hst.Addrs())
				if nil != err {
					fmt.Println("connect fail")
				}else {
					fmt.Printf("conn[%d] succeed !\n", idx)
				}
			}(&connSilce[idx], &hstcSilce[idx])
		}
		wg.Wait()
	}

	//--------------------------------------------------

	for index, conn := range connSilce {
		idx := index
		go func() {
			if conn == nil {
				return
			}
			for {
				ctx_local, cancel_local := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel_local()
				_, err := conn.SendMsg(ctx_local, 0x0, []byte("111111111"))
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Printf("conn[%d] send succeed\n", idx)
					//time.Sleep(time.Second * 2)
				}
				time.Sleep(time.Second * 2)
			}
		}()
	}

	for {
		time.Sleep(time.Second*1)
	}
}