package host_test

import (
	"context"
	ra "crypto/rand"
	"fmt"
	ic "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/mr-tron/base58"
	"github.com/multiformats/go-multiaddr"
	host "github.com/yottachain/YTHost"
	ci "github.com/yottachain/YTHost/clientInterface"
	. "github.com/yottachain/YTHost/hostInterface"
	"github.com/yottachain/YTHost/option"
	"github.com/yottachain/YTHost/service"
	"math/rand"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// 测试创建通讯节点
func TestNewHost(t *testing.T) {
	var localMa2 = "/ip4/0.0.0.0/tcp/9003"

	pk, _, _ := ic.GenerateSecp256k1Key(ra.Reader) //option Identity cover
	ma, _ := multiaddr.NewMultiaddr(localMa2)      // option listenaddr cover
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
	_, err := hst2.Connect(ctx, hst.Config().ID, hst.Addrs())
	if err != nil {
		t.Fatal(err.Error())
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

	peerInfo, _ := clt.RemotePeer()

	// 打印节点信息
	t.Log(peerInfo.ID.Pretty(), peerInfo.Addrs)
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
	if res, err := clt.SendMsg(context.Background(), hst.Config().ID, 0x11, []byte("2222")); err != nil {
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
	if res, err := clt.SendMsg(context.Background(), hst.Config().ID, 0x11, []byte("2222")); err != nil {
		t.Fatal(err)
	} else {
		t.Log(string(res))
	}
	if res, err := clt.SendMsg(context.Background(), hst.Config().ID, 0x12, []byte("121212")); err != nil {
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
	if res, err := clt.SendMsgClose(context.Background(), hst.Config().ID, 0x11, []byte("2222")); err != nil {
		t.Fatal(err)
	} else {
		t.Log(string(res))
	}
	if res, err := clt.SendMsg(context.Background(), hst.Config().ID, 0x12, []byte("121212")); err == nil {
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
	if res, err := clt.SendMsgClose(context.Background(), hst.Config().ID, 0x11, []byte("2222")); err != nil {
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

	go hst2.Accept()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err := hst2.ClientStore().Get(ctx, hst.Config().ID, hst.Addrs())
	if err != nil {
		t.Fatal(err)
	}

	// 第二次获取相同连接，不应建立新的连接
	_, err = hst2.ClientStore().Get(ctx, hst.Config().ID, hst.Addrs())
	if err != nil {
		t.Fatal(err)
	}
	/*if hst2.ClientStore().Len() > 1 {
		t.Fatal("不应该建立新连接")
	}*/

	// 另一种发送消息的方式
	if res, err := hst2.SendMsg(context.Background(), hst.Config().ID, 0x11, []byte("2222")); err != nil {
		t.Fatal(err)
	} else {
		t.Log(string(res))
	}
	/*if err := hst2.ClientStore().Close(hst.Config().ID); err != nil {
		t.Fatal(err)
	}*/
	/*if res, err := clt.SendMsg(context.Background(), 0x12, []byte("121213")); err == nil {
		t.Fatal(fmt.Errorf("此处连接应该关闭"), res)
	} else {
		t.Log(err)
	}*/
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
		count = count + 1
		//fmt.Println(fmt.Sprintf("msg num is %d", count), string(requestData))
		fmt.Println(fmt.Sprintf("msg num is %d", count))
		return []byte("111111111111"), nil
	})

	go hst.Accept()

	lenth := 5000
	hstcSilce := make([]Host, lenth)
	connSilce := make([]ci.YTHClient, lenth)
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
		} else {
			fmt.Println(connSilce[j].LocalPeer())
		}

		ctx1, cancel1 := context.WithCancel(context.Background())
		defer cancel1()
		go func(ctx context.Context, conn *ci.YTHClient, index int) (error) {
			for {
				select {
				case <-ctx.Done():
					return nil
				default:
					ctx_local, cancel_local := context.WithTimeout(context.Background(), time.Second*10)
					defer cancel_local()
					_, err := (*conn).SendMsg(ctx_local, hstcSilce[j].Config().ID, 0x0, []byte("111111111"))
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
		time.Sleep(time.Second * 1)
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
		count = count + 1
		//fmt.Println(fmt.Sprintf("msg num is %d", count), string(requestData))
		fmt.Println(fmt.Sprintf("msg num is %d", count))
		return []byte("111111111111"), nil
	})

	go hst.Accept()

	lenth := 5000
	//hstcSilce := make([]Host, lenth)
	connSilce := make([]ci.YTHClient, lenth)

	for k, _ := range connSilce {
		ma, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", 13000+k))
		h, err := host.NewHost(option.ListenAddr(ma))
		if err != nil {
			fmt.Println(err)
		} else {
			conn, _ := h.Connect(context.Background(), hst.Config().ID, hst.Addrs())
			connSilce[k] = conn
		}
		time.After(time.Millisecond * 100)
	}

	for {
		wg := sync.WaitGroup{}
		wg.Add(lenth)
		for _, c := range connSilce {
			go func() {
				if c == nil {
					return
				}
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				c.SendMsg(ctx, hst.Config().ID, 0x0, []byte("msg"))
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
		count = count + 1
		//fmt.Println(fmt.Sprintf("msg num is %d", count), string(requestData))
		fmt.Println(fmt.Sprintf("msg num is %d", count))
		return []byte("111111111111"), nil
	})

	go hst.Accept()

	lenth := 50
	hstcSilce := make([]Host, lenth)
	connSilce := make([]ci.YTHClient, lenth)

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
	concurNum := 5
	if concurNum > lenth {
		concurNum = lenth
	}
	for i := 0; i < lenth; i = i + concurNum {
		curEndPos := i + concurNum
		stepSize := 0
		wg := sync.WaitGroup{}
		if curEndPos < lenth {
			stepSize = concurNum
		} else {
			stepSize = lenth - i
		}
		wg.Add(stepSize)

		for j := i; j < i+stepSize; j++ {
			idx := j
			go func(conn *ci.YTHClient, host *Host) {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
				defer cancel()
				*conn, err = (*host).ClientStore().Get(ctx, hst.Config().ID, hst.Addrs())
				if nil != err {
					fmt.Printf("conn[%d] fail !!!\n", idx)
					return
				} else {
					fmt.Printf("conn[%d] succeed !\n", idx)
				}
				// 立即发送一条消息
				_, err := (*conn).SendMsg(ctx, hst.Config().ID, 0x0, []byte("111111111"))
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Printf("conn[%d] send succeed\n", idx)
					//time.Sleep(time.Second * 2)
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
				_, err := conn.SendMsg(ctx_local, hst.Config().ID, 0x0, []byte("111111111"))
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
		time.Sleep(time.Second * 1)
	}
}

func TestSendTimeOut(t *testing.T) {
	mastr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", 9000)
	ma, _ := multiaddr.NewMultiaddr(mastr)
	/*pristr, _ := YTCrypto.CreateKey()
	privbytes, err := base58.Decode(pristr)
	if err != nil {
		panic(err)
	}
	pk, err := ic.UnmarshalSecp256k1PrivateKey(privbytes[1:33])
	if err != nil {
		panic(err)
	}
	hst, err := host.NewHost(option.ListenAddr(ma), option.Identity(pk))*/
	hst, err := host.NewHost(option.ListenAddr(ma))
	if err != nil {
		panic(err)
	}

	var count int64 = 0

	hst.RegisterGlobalMsgHandler(func(requestData []byte, head service.Head) (bytes []byte, e error) {
		count = count + 1
		fmt.Println(fmt.Sprintf("msg num is %d", count))
		//time.Sleep(time.Second*220)
		fmt.Println(fmt.Sprintf("msg return..................."))
		return []byte("111111111111"), nil
	})

	go hst.Accept()

	hst2 := GetRandomHost()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	clt, err := hst2.Connect(ctx, hst.Config().ID, hst.Addrs())
	// 关闭连接
	defer clt.Close()
	if err != nil {
		t.Fatal(err.Error())
	}
	if res, err := clt.SendMsg(context.Background(), hst.Config().ID, 0x0, []byte("22222222223333333333333333")); err != nil {
		t.Fatal(err)
	} else {
		fmt.Println(string(res))
	}

	/*for {
		time.Sleep(time.Second*1)
	}*/
}

func TestStressConn3(t *testing.T) {
	const max = 30000
	//var number = 0
	net.Listen("tcp4", "127.0.0.1:9003")
	wg := sync.WaitGroup{}
	wg.Add(max)

	/*go func() {
		for {
			number++
			fmt.Println("accept", time.Now().String())
			//l.Accept()
			fmt.Println("end", time.Now().String(), number)
			wg.Done()
		}
	}()*/

	for i := 0; i < max; i++ {
		go func() {
			t.Log("start conn")
			net.Dial("tcp4", "127.0.0.1:9003")
			t.Log("end conn")
			wg.Done()
		}()
	}

	tm := time.Now()
	wg.Wait()
	t.Log("用时", time.Now().Sub(tm).Seconds())
}

//测试多连接消息加密
func TestMutlConnCrypt(t *testing.T) {
	mastr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", 9000)

	ma, _ := multiaddr.NewMultiaddr(mastr)
	hst, err := host.NewHost(option.ListenAddr(ma))
	if err != nil {
		panic(err)
	}

	var count int64 = 0

	hst.RegisterGlobalMsgHandler(func(requestData []byte, head service.Head) (bytes []byte, e error) {
		count = count + 1
		fmt.Println(fmt.Sprintf("msg num is %d, msg is [%s]", count, string(requestData)))
		return []byte("receice----------------------succeed!"), nil
	})

	go hst.Accept()

	lenth := 1
	hstcSilce := make([]Host, lenth)
	connSilce := make([]ci.YTHClient, lenth)

	for i := 0; i < lenth; i++ {
		j := i
		port := 19000 + j + 1
		mastr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)

		ma, _ := multiaddr.NewMultiaddr(mastr)
		hstcSilce[j], err = host.NewHost(option.ListenAddr(ma))
		go hstcSilce[j].Accept()
		if err != nil {
			hstcSilce[j] = nil
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		/*mastr = "/ip4/127.0.0.1/tcp/9000/p2p/" + hst.Config().ID.String() + "/p2p-circuit/"
		maddrs := make([]multiaddr.Multiaddr, 1)
		maddrs[0], err = multiaddr.NewMultiaddr(mastr)
		connSilce[j], err = hstcSilce[j].Connect(ctx, hst.Config().ID, maddrs)*/

		connSilce[j], err = hstcSilce[j].Connect(ctx, hst.Config().ID, hst.Addrs())
		if err != nil {
			t.Log(err.Error())
			continue
		} else {
			//fmt.Println(connSilce[j].LocalPeer())
		}
	}

	for index, conn := range connSilce {
		if nil != conn {
			ctx_local, cancel_local := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel_local()

			key, _, _ := ic.GenerateSecp256k1Key(ra.Reader)
			keyByte, _ := key.Raw()
			msgstr := base58.Encode(keyByte)

			//fmt.Println(keystr)		//发送的消息

			ret, err := (conn).SendMsg(ctx_local, hst.Config().ID, 0x0, []byte(msgstr))
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("conn[%d] send succeed, msg is [%s]\n", index, msgstr)
				fmt.Println(string(ret))
			}

			time.Sleep(time.Second * 5)
		}
	}

	/*for {
		time.Sleep(time.Second*1)
	}*/
}

func TestStrSplit(t *testing.T) {
	str := "/ip4/152.136.198.134/tcp/9001/p2p/16Uiu2HAm3txKmVThw2xVge7vqV2zhev4nd1m2Tt9cQV1pbUxh4HJ/p2p-circuit/"
	//str := "/ip4/152.136.198.134/tcp/9001/p2p/"
	strSplics := strings.Split(str, "/")
	dStr := "/"
	lenth := len(strSplics)
	indx := 0
	for i, v := range strSplics {
		indx = i
		if "" != v {
			if "p2p" == v {
				break
			}
			fmt.Println(strSplics[i])
			dStr = dStr + v + "/"
		}
	}
	if lenth != indx+1 {
		fmt.Println(strSplics[indx+1])
	}
	//dStr = "/" + dStr
	fmt.Println(dStr)
}

func TestRelayTransMsg(t *testing.T) {
	mastr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", 9000)
	ma, _ := multiaddr.NewMultiaddr(mastr)
	hst1, err := host.NewHost(option.ListenAddr(ma), option.OpenDebug(), option.OpenPProf("127.0.0.1:8888"))
	if err != nil {
		panic(err)
	}
	go hst1.Accept()

	mastr = fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", 9001)
	ma, _ = multiaddr.NewMultiaddr(mastr)
	hst2, err := host.NewHost(option.ListenAddr(ma), option.OpenDebug(), option.OpenPProf("127.0.0.1:8889"))
	if err != nil {
		panic(err)
	}
	go hst2.Accept()

	mastr = fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", 9002)
	ma, _ = multiaddr.NewMultiaddr(mastr)
	hst3, err := host.NewHost(option.ListenAddr(ma), option.OpenDebug())
	if err != nil {
		panic(err)
	}
	go hst3.Accept()

	hst3.RegisterGlobalMsgHandler(func(requestData []byte, head service.Head) (bytes []byte, e error) {
		fmt.Println(fmt.Sprintf("msg is [%s]\n", string(requestData)))
		return []byte("receice----hst3-----hst3------hst3-------succeed!"), nil
	})

	hst1.RegisterGlobalMsgHandler(func(requestData []byte, head service.Head) (bytes []byte, e error) {
		fmt.Println(fmt.Sprintf("msg is [%s]", string(requestData)))
		return []byte("receice----hst1------hst1------hst1------succeed!"), nil
	})

	fmt.Println(hst1.Config().ID.String())
	fmt.Println(hst2.Config().ID.String())
	fmt.Println(hst3.Config().ID.String())
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	_, err = hst1.ClientStore().Get(ctx, hst2.Config().ID, hst2.Addrs())
	if nil != err {
		panic(err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	_, err = hst3.ClientStore().Get(ctx, hst2.Config().ID, hst2.Addrs())
	if nil != err {
		panic(err)
	}

	mastr = "/ip4/127.0.0.1/tcp/9001/p2p/" + hst2.Config().ID.String() + "/p2p-circuit/"
	maddrs := make([]multiaddr.Multiaddr, 1)
	maddrs[0], err = multiaddr.NewMultiaddr(mastr)
	ctx, cancel = context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	var cltt1 *ci.YTHClient
	clt1, err := hst1.ClientStore().Get(ctx, hst3.Config().ID, maddrs)
	cltt1 = &clt1
	if nil != err {
		panic(err)
	}

	mastr = "/ip4/127.0.0.1/tcp/9001/p2p/" + hst2.Config().ID.String() + "/p2p-circuit/"
	maddrs = make([]multiaddr.Multiaddr, 1)
	maddrs[0], err = multiaddr.NewMultiaddr(mastr)
	ctx, cancel = context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	clt2, err := hst3.ClientStore().Get(ctx, hst1.Config().ID, maddrs)
	if nil != err {
		panic(err)
	}

	go func() {
		time.Sleep(time.Second * 60)
		_ = hst1.ClientStore().Close(hst3.Config().ID)
		cltt1 = nil
		time.Sleep(time.Second * 60)
		mastr = "/ip4/127.0.0.1/tcp/9001/p2p/" + hst2.Config().ID.String() + "/p2p-circuit/"
		maddrs := make([]multiaddr.Multiaddr, 1)
		maddrs[0], err = multiaddr.NewMultiaddr(mastr)

		time.Sleep(time.Second * 3)
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()
		clt1, err = hst1.ClientStore().Get(ctx, hst3.Config().ID, maddrs)
		if err != nil {
			fmt.Println(err)
		} else {
			cltt1 = &clt1
		}
	}()

	hst3ID := hst3.Config().ID
	hst1ID := hst1.Config().ID
	for i := 0; i < 1000; i++ {
		ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel1()
		msg1 := []byte("host1----->>>host2-------->>>host3  d22222222askfjpqejwogeoqnbqdkjdkjklajfdljfalksjfkjdsaklfjlkdajsfkl3333333djslkfjasjfladsjlfj\n" +
			"vczvlkcjxvljkzcl;xjvk;cjv;lzkcxj;lvkfgagsdgggggggggggggggggggggggggggggggggggggvcbxbbxzcz3333bxc\n" +
			"vczvlkcjxvljkzcl;xjvk;cjv;lzkcxj;lvkfgagsdgggggggggggggggggggggggggggggggggggggvcbxbbxzczbxc\n" +
			"daskfjpqejwogeoqnbqdkjdkjklajfdljfalksjfkjdsaklfjlkdajsfkldjslkfjasjfladsjlfj111111111111ddddddddddd\n" +
			"d22222222askfjpqejwogeoqnbqdkjdkjklajfdljfalksjfkjdsaklfjlkdajsfkl3333333djslkfjasjfladsjlfj\n" +
			"vczvlkcjxvljkzcl;xjvk;cjv;lzkcxj;lvkfgagsdgggggggggggggggggggggggggggggggggggggvcbxbbxzcz3333bxc\n" +
			"vczvlkcjxvljkzcl;xjvk;cjv;lzkcxj.;lvkfgagsdgggggggggggggggggggggggggggggggggggggvcbxbbxzczbxc\n" +
			"daskfjpqejwogeoqnbqdkjdkjklajfdljfalksjfkjdsaklfjlkdajsfkldjslkfjasjfladsjlfj111111111111ddddddddddd")
		if (cltt1 != nil) && (*cltt1) != nil && (!(*cltt1).IsClosed()) {
			ret, err := (*cltt1).SendMsg(ctx1, hst3ID, 0x0, msg1)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(string(ret))
			}
		}

		ctx1, cancel1 = context.WithTimeout(context.Background(), time.Second*60)
		defer cancel1()
		msg2 := []byte("host3----->>>host2-------->>>host1   1111111111111111111111119999999999999999999999999dgggggggggggggggggggggggggggggggggggggvcbxbbxzcz3333bxc\n" +
			"vczvlkcjxvljkzcl;xjvk;cjvlksjfkjdsaklfjlkdajsfkl3333333d;lzkcxj;lvkfgagsdgsssssssssssssssssgvcbxbbxzczbxc\n" +
			"daskfjpqejwogeoqnbqdkjdkjksssssssssssssssssssssssssslkfjasjfladsjlfj111111111111ddddddddddd\n" +
			"d22222222askf555555555555555555555555555533djslkfjasjfladsjlfj\n" +
			"vczvlkcjxvljkzcl;xjvk;cjv;lzkcxj;lvkfgagsdgg666666666666666666666666666666gggvcbxbbxzcz3333bxc\n" +
			"vczvlkcjxvljkzcl;xjvk;cjv;lzkcxj;lvkfga7777777777777777777777777gggggvcbxbbxzczbxc\n" +
			"daskfjpqejwoge8888888888888888888888888888888888888888888888888888st2---- ni hao xiao huo ban")
		if !clt2.IsClosed() {
			ret, err := clt2.SendMsg(ctx1, hst1ID, 0x0, msg2)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(string(ret))
			}
		}

		time.Sleep(time.Second * 2)
	}

	for {
		time.Sleep(time.Second * 60)
		break
	}

}

func TestRelayTransMsg1(t *testing.T) {
	mastr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", 9000)
	ma, _ := multiaddr.NewMultiaddr(mastr)
	hst1, err := host.NewHost(option.ListenAddr(ma), option.OpenPProf("127.0.0.1:8888"))
	if err != nil {
		panic(err)
	}
	go hst1.Accept()

	mastr = fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", 9001)
	ma, _ = multiaddr.NewMultiaddr(mastr)
	hst2, err := host.NewHost(option.ListenAddr(ma), option.OpenPProf("127.0.0.1:8889"))
	if err != nil {
		panic(err)
	}
	go hst2.Accept()

	hst2.RegisterGlobalMsgHandler(func(requestData []byte, head service.Head) (bytes []byte, e error) {
		fmt.Println(fmt.Sprintf("msg is [%s]", string(requestData)))
		return []byte("receice----hst2------hst2------hst2------succeed!"), nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	clt, err := hst1.ClientStore().Get(ctx, hst2.Config().ID, hst2.Addrs())
	if nil != err {
		panic(err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	_, err = hst2.ClientStore().Get(ctx, hst1.Config().ID, hst1.Addrs())
	//_, _, err = hst2.Connect(ctx, hst1.Config().ID, hst1.Addrs())
	if nil != err {
		panic(err)
	}

	for i := 0; i < 1000; i++ {
		ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel1()
		msg1 := []byte("host1----->>>host2  d22222222askfjpqejwogeoqnbqdkjdkjklajfdljfalksjfkjdsaklfjlkdajsfkl3333333djslkfjasjfladsjlfj\n" +
			"vczvlkcjxvljkzcl;xjvk;cjv;lzkcxj;lvkfgagsdgggggggggggggggggggggggggggggggggggggvcbxbbxzcz3333bxc\n" +
			"vczvlkcjxvljkzcl;xjvk;cjv;lzkcxj;lvkfgagsdgggggggggggggggggggggggggggggggggggggvcbxbbxzczbxc\n" +
			"daskfjpqejwogeoqnbqdkjdkjklajfdljfalksjfkjdsaklfjlkdajsfkldjslkfjasjfladsjlfj111111111111ddddddddddd\n" +
			"d22222222askfjpqejwogeoqnbqdkjdkjklajfdljfalksjfkjdsaklfjlkdajsfkl3333333djslkfjasjfladsjlfj\n" +
			"vczvlkcjxvljkzcl;xjvk;cjv;lzkcxj;lvkfgagsdgggggggggggggggggggggggggggggggggggggvcbxbbxzcz3333bxc\n" +
			"vczvlkcjxvljkzcl;xjvk;cjv;lzkcxj;lvkfgagsdgggggggggggggggggggggggggggggggggggggvcbxbbxzczbxc\n" +
			"daskfjpqejwogeoqnbqdkjdkjklajfdljfalksjfkjdsaklfjlkdajsfkldjslkfjasjfladsjlfj111111111111ddddddddddd")

		ret, err := clt.SendMsg(ctx1, hst2.Config().ID, 0x0, msg1)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(string(ret))
		}

		time.Sleep(time.Second * 2)
	}

	for {
		time.Sleep(time.Second * 60)
		break
	}

}

func TestRelayTransMsg2(t *testing.T) {
	mastr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", 9000)
	ma, _ := multiaddr.NewMultiaddr(mastr)
	hst1, err := host.NewHost(option.ListenAddr(ma), option.OpenPProf("127.0.0.1:8888"))
	if err != nil {
		panic(err)
	}
	go hst1.Accept()

	mastr = fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", 9001)
	ma, _ = multiaddr.NewMultiaddr(mastr)
	hst2, err := host.NewHost(option.ListenAddr(ma), option.OpenPProf("127.0.0.1:8889"))
	if err != nil {
		panic(err)
	}
	go hst2.Accept()

	maddrs := hst2.Addrs()
	addrs := make([]string, len(maddrs))
	for k, m := range maddrs {
		addrs[k] = m.String()
	}
	for _, addr := range addrs {
		fmt.Println(addr)
	}

	mastr = fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", 9002)
	ma, _ = multiaddr.NewMultiaddr(mastr)
	hst3, err := host.NewHost(option.ListenAddr(ma))
	if err != nil {
		panic(err)
	}
	go hst3.Accept()

	fmt.Println(hst1.Config().ID.String())
	fmt.Println(hst2.Config().ID.String())
	fmt.Println(hst3.Config().ID.String())
	fmt.Println()

	hst2.RegisterGlobalMsgHandler(func(requestData []byte, head service.Head) (bytes []byte, e error) {
		fmt.Println(fmt.Sprintf("msg is [%s]", string(requestData)))
		return []byte("receice----hst2------hst2------hst2------succeed!"), nil
	})

	//mastr = "/ip4/127.0.0.1/tcp/9001/"
	mastr = "/ip4/192.168.3.182/tcp/9001/"
	//mastr = "/ip4/172.17.32.8/tcp/9001/"
	maddrs = make([]multiaddr.Multiaddr, 1)
	maddrs[0], err = multiaddr.NewMultiaddr(mastr)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	//clt1, err := hst1.ClientStore().Get(ctx, hst2.Config().ID, hst2.Addrs())
	clt1, err := hst1.ClientStore().Get(ctx, hst2.Config().ID, maddrs)
	if nil != err {
		panic(err)
	}
	hst1.ClientStore().PrintConnInfo(clt1)
	fmt.Println("---------------------------")

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	clt2, err := hst3.ClientStore().Get(ctx, hst2.Config().ID, hst2.Addrs())
	if nil != err {
		panic(err)
	}
	hst3.ClientStore().PrintConnInfo(clt2)
	fmt.Println("---------------------------")

	//mastr = "/ip4/127.0.0.1/tcp/9001/p2p/" + hst2.Config().ID.String() + "/p2p-circuit/"
	mastr = "/ip4/192.168.3.182/tcp/9001/p2p/" + hst2.Config().ID.String() + "/p2p-circuit/"
	//mastr = "/ip4/172.17.32.8/tcp/9001/p2p/" + hst2.Config().ID.String() + "/p2p-circuit/"
	maddrs = make([]multiaddr.Multiaddr, 1)
	maddrs[0], err = multiaddr.NewMultiaddr(mastr)
	ctx, cancel = context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	clt3, err := hst1.ClientStore().Get(ctx, hst3.Config().ID, maddrs)

	if nil != err {
		panic(err)
	}

	hst1.ClientStore().PrintConnInfo(clt3)
	fmt.Println("---------------------------")

	_ = hst1.ClientStore().Close(hst2.Config().ID)
	hst1.ClientStore().PrintConnInfo(clt3)
	fmt.Println("---------------------------")

	_ = hst1.ClientStore().Close(hst3.Config().ID)
	hst1.ClientStore().PrintConnInfo(clt3)
	fmt.Println("---------------------------")

	mastr = "/ip4/192.168.3.182/tcp/9001/p2p/" + hst2.Config().ID.String() + "/p2p-circuit/"
	//mastr ="/ip4/172.17.32.8/tcp/9001/p2p/" + hst2.Config().ID.String() + "/p2p-circuit/"
	maddrs = make([]multiaddr.Multiaddr, 1)
	maddrs[0], err = multiaddr.NewMultiaddr(mastr)

	for i := 0; i < 1000; i++ {
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*600)
		defer cancel()
		_, err = hst1.ClientStore().Get(ctx, hst3.Config().ID, maddrs)

		if nil != err {
			time.Sleep(time.Second * 1)
			//panic(err)
			fmt.Println("try again")
		} else {
			fmt.Println("connect succeed")
			break
		}
	}
	if nil != err {
		panic(err)
	}

	_ = hst3.ClientStore().Close(hst2.Config().ID)
	_ = hst1.ClientStore().Close(hst3.Config().ID)

	for {
		time.Sleep(time.Second * 60)
		break
	}

}

func TestRelayTransMsg3(t *testing.T) {
	mastr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", 9000)
	ma, _ := multiaddr.NewMultiaddr(mastr)
	hst1, err := host.NewHost(option.ListenAddr(ma), option.OpenPProf("127.0.0.1:8888"))
	if err != nil {
		panic(err)
	}
	go hst1.Accept()

	mastr = fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", 9001)
	ma, _ = multiaddr.NewMultiaddr(mastr)
	hst2, err := host.NewHost(option.ListenAddr(ma), option.OpenDebug(), option.OpenPProf("127.0.0.1:8889"))
	if err != nil {
		panic(err)
	}
	go hst2.Accept()

	hst1.RegisterGlobalMsgHandler(func(requestData []byte, head service.Head) (bytes []byte, e error) {
		fmt.Println(fmt.Sprintf("msg is [%s]", string(requestData)))
		return []byte("receice----hst1------hst1------hst1------succeed!"), nil
	})

	hst2.RegisterGlobalMsgHandler(func(requestData []byte, head service.Head) (bytes []byte, e error) {
		fmt.Println(fmt.Sprintf("msg is [%s]", string(requestData)))
		return []byte("receice----hst2------hst2------hst2------succeed!"), nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	clt1, err := hst1.ClientStore().Get(ctx, hst2.Config().ID, hst2.Addrs())
	if nil != err {
		panic(err)
	}

	ctx2, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	//clt2, rid, err := hst2.Connect(ctx2, hst1.Config().ID, hst1.Addrs())
	clt2, err := hst2.ClientStore().Get(ctx2, hst1.Config().ID, hst1.Addrs())
	if nil != err {
		panic(err)
	} else {
		//fmt.Println(rid.String()+"aaaaaaaaaaaaaaa")
	}

	msg1 := []byte("1111111111111111111111111")
	msg2 := []byte("22222222222222222222222222")
	for i := 0; i < 1000; i++ {
		ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel1()
		ret, err := clt1.SendMsg(ctx1, hst2.Config().ID, 0x0, msg1)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(string(ret))
		}
		time.Sleep(time.Second * 1)
		ret, err = clt2.SendMsg(ctx1, hst1.Config().ID, 0x0, msg2)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(string(ret))
		}
		time.Sleep(time.Second * 1)
	}

	for {
		time.Sleep(time.Second * 60)
		break
	}

}

func TestRelayTransMsg4(t *testing.T) {
	mastr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", 9000)
	ma, _ := multiaddr.NewMultiaddr(mastr)
	hst1, err := host.NewHost(option.ListenAddr(ma), option.OpenPProf("127.0.0.1:8888"))
	if err != nil {
		panic(err)
	}
	go hst1.Accept()

	mastr = fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", 9001)
	ma, _ = multiaddr.NewMultiaddr(mastr)
	hst2, err := host.NewHost(option.ListenAddr(ma), option.OpenPProf("127.0.0.1:8889"))
	if err != nil {
		panic(err)
	}
	go hst2.Accept()

	maddrs := hst2.Addrs()
	addrs := make([]string, len(maddrs))
	for k, m := range maddrs {
		addrs[k] = m.String()
	}
	for _, addr := range addrs {
		fmt.Println(addr)
	}

	fmt.Println(hst1.Config().ID.String())
	fmt.Println(hst2.Config().ID.String())

	hst2.RegisterGlobalMsgHandler(func(requestData []byte, head service.Head) (bytes []byte, e error) {
		fmt.Println(fmt.Sprintf("msg is [%s]", string(requestData)))
		return []byte("receice----hst2------hst2------hst2------succeed!"), nil
	})

	//mastr = "/ip4/127.0.0.1/tcp/9001/"
	mastr = "/ip4/192.168.3.182/tcp/9001/"
	//mastr = "/ip4/172.17.32.8/tcp/9001/"
	maddrs = make([]multiaddr.Multiaddr, 1)
	maddrs[0], err = multiaddr.NewMultiaddr(mastr)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*600)
	defer cancel()
	clt1, err := hst1.ClientStore().Get(ctx, hst2.Config().ID, hst2.Addrs())
	//clt1, err := hst1.ClientStore().Get(ctx, hst2.Config().ID, maddrs)
	if nil != err {
		panic(err)
	}
	hst1.ClientStore().PrintConnInfo(clt1)
	fmt.Println("---------------------------")

	_ = hst1.ClientStore().Close(hst2.Config().ID)
	hst1.ClientStore().PrintConnInfo(clt1)
	fmt.Println("---------------------------")

	mastr = "/ip4/192.168.3.182/tcp/9001/"
	//mastr = "/ip4/172.17.32.8/tcp/9001/"
	maddrs = make([]multiaddr.Multiaddr, 1)
	maddrs[0], err = multiaddr.NewMultiaddr(mastr)

	for i := 0; i < 1000; i++ {
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*6)
		defer cancel()
		//_, err = hst1.ClientStore().Get(ctx, hst2.Config().ID, maddrs)
		_, err = hst1.ClientStore().Get(ctx, hst2.Config().ID, hst2.Addrs())

		if nil != err {
			time.Sleep(time.Second * 1)
			//panic(err)
			fmt.Println("try again")
		} else {
			fmt.Println("connect succeed")
			break
		}
	}
	if nil != err {
		panic(err)
	}

	for {
		//time.Sleep(time.Second * 60)
		break
	}

}
