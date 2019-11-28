package host_test

import (
	"context"
	"fmt"
	host "github.com/graydream/YTHost"
	"github.com/graydream/YTHost/option"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
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
		maddrs, _ := hst.Addrs()
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

func TestMnet(t *testing.T) {
	var localMa = "/ip4/0.0.0.0/tcp/9010"
	//var localMa2 = "/ip4/0.0.0.0/tcp/9012"

	ma, _ := multiaddr.NewMultiaddr(localMa)
	//ma2, _ := multiaddr.NewMultiaddr(localMa2)

	ls, err := manet.Listen(ma)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		for {
			conn, err := ls.Accept()
			fmt.Println(conn.RemoteAddr(), err)
		}
	}()

	conn, err := manet.Dial(ma)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(conn.RemoteAddr())

}

func TestConn(t *testing.T) {
	var localMa = "/ip4/0.0.0.0/tcp/9011"
	var localMa2 = "/ip4/0.0.0.0/tcp/9012"

	ma, _ := multiaddr.NewMultiaddr(localMa)
	ma2, _ := multiaddr.NewMultiaddr(localMa2)

	hst, err := host.NewHost(option.ListenAddr(ma))
	if err != nil {
		t.Fatal(err.Error())
	}
	if err := hst.Server().Register(&RpcService{}); err != nil {
		t.Fatal(err)
	}

	go hst.Accept()

	hst2, err := host.NewHost(option.ListenAddr(ma2))
	if err != nil {
		t.Fatal(err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	//addrs, _ := hst.Addrs()
	clt, err := hst2.Connect(ctx, hst.Config().ID, []multiaddr.Multiaddr{ma})
	if err != nil {
		t.Fatal(err.Error())
	}
	var res Reply
	if err := clt.Call("rpcService.Ping", "", &res); err != nil {
		t.Fatal(err)
	} else {
		t.Log(res.Value)
	}
}

// 测试发送protobuf消息。注册消息处理器
func TestConnSendProtobufMsgAndHandler(t *testing.T) {
}
