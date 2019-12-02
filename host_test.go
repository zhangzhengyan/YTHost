package host_test

import (
	"context"
	"fmt"
	host "github.com/graydream/YTHost"
	"github.com/graydream/YTHost/option"
	"github.com/multiformats/go-multiaddr"
	"math/rand"
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

func GetRandomHost() host.Host {
	hst, err := host.NewHost(option.ListenAddr(GetRandomLocalMutlAddr()))
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
	t.Log(peerInfo.ID.Pretty(), peerInfo.Addrs)
}

// 发送，处理消息
func TestHandleMsg(t *testing.T) {
	hst := GetRandomHost()
	if err := hst.Server().Register(new(RpcService)); err != nil {
		t.Fatal(err)
	}

	hst.RegisterHandler(0x11, func(requestData []byte) (bytes []byte, e error) {
		t.Log(string(requestData))
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
