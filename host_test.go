package host_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	host "github.com/graydream/YTHost"
	. "github.com/graydream/YTHost/hostInterface"
	"github.com/graydream/YTHost/option"
	"github.com/graydream/YTHost/service"
	"github.com/multiformats/go-multiaddr"
	"math/rand"
	"sync"
	"testing"
	"time"
)

// 测试创建通讯节点
func TestNewHost(t *testing.T) {
	var localMa2 = "/ip4/0.0.0.0/tcp/9003"

	ma, _ := multiaddr.NewMultiaddr(localMa2)
	if hst, err := host.NewHost(option.ListenAddr(ma)); err != nil {
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

func Ping(h Host, i int) bool {
	ma, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", i+10000))
	h2 := GetHost(ma)
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancle()
	fmt.Println("开始连接", i)
	clt, err := h2.Connect(ctx, h.Config().ID, h.Addrs())
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer clt.Close()
	res, err := clt.SendMsg(ctx, 0x13, []byte{1})
	if err != nil {
		fmt.Println(err)
		return false
	} else if string(res) != "pong" {
		fmt.Println("error", string(res))
		return false
	}
	return true
}

func TestTimeout(t *testing.T) {
	h := GetRandomHost()
	//go h.Accept()
	h2 := GetRandomHost()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if clt, err := h2.Connect(ctx, h.Config().ID, h.Addrs()); err != nil {
		t.Fatal(err)
	} else {
		_, err := clt.SendMsg(ctx, 0x13, []byte{12})
		t.Log(err)
		//if clt.Ping(ctx) {
		//	t.Log("ping success")
		//} else {
		//	t.Log("time out")
		//}
	}
}

func TestStress(t *testing.T) {
	h1 := GetRandomHost()
	_ = h1.RegisterHandler(0x13, func(requestData []byte, head service.Head) (bytes []byte, e error) {
		return []byte("pong"), nil
	})
	go h1.Accept()
	errCount := 0
	successCount := 0
	const max_count = 20000
	q := make(chan struct{}, 10000)
	wg := sync.WaitGroup{}
	wg.Add(max_count)
	for i := 0; i < max_count; i++ {
		go func(i int) {
			q <- struct{}{}
			if Ping(h1, i) {
				successCount++
				//fmt.Println("success", i)
			} else {
				errCount++
				//fmt.Println("error", i)
			}
			defer wg.Done()
			defer func() { <-q }()
		}(i)
	}
	wg.Wait()

	t.Log("success count", successCount)
	t.Log("error count", errCount)
}

func TestBytes(t *testing.T) {
	const x int32 = 0x110
	buf := bytes.NewBuffer([]byte{})
	err := binary.Write(buf, binary.BigEndian, x)
	fmt.Println(err)
	var x2 int32
	err = binary.Read(buf, binary.LittleEndian, &x2)
	fmt.Println(err)
	fmt.Printf("%x", x2)
}

func TestCtx(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*3500)
	defer cancel()
	go func(ctx context.Context) {
		select {
		case <-ctx.Done():
			fmt.Println(1)
			return
		default:
			i := 0
			for {
				i++
				fmt.Println("i:", i)
				<-time.After(time.Second)
			}
		}
	}(ctx)
	go func(ctx context.Context) {
		i2 := 0
		for {
			select {
			case <-ctx.Done():
				fmt.Println(2)
				return
			default:
				i2++
				fmt.Println("i2:", i2)
				<-time.After(time.Second)
			}
		}
	}(ctx)
	select {
	case <-ctx.Done():
		fmt.Println(3)
		//return
	}
	select {}
}
