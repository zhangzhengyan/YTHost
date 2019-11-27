package host_test

import (
	"context"
	"fmt"
	"github.com/gogo/protobuf/proto"
	host "github.com/graydream/YTHost"
	"github.com/graydream/YTHost/option"
	"github.com/graydream/YTHost/pb"
	"github.com/graydream/YTHost/pbMsgHandler"
	"github.com/multiformats/go-multiaddr"
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

func TestConn(t *testing.T) {
	var localMa = "/ip4/0.0.0.0/tcp/9010"
	var localMa2 = "/ip4/0.0.0.0/tcp/9012"

	ma, _ := multiaddr.NewMultiaddr(localMa)
	ma2, _ := multiaddr.NewMultiaddr(localMa2)

	hst, err := host.NewHost(option.ListenAddr(ma))
	go hst.Serve(context.Background())
	if err != nil {
		t.Fatal(err.Error())
	}
	hst2, err := host.NewHost(option.ListenAddr(ma2))
	if err != nil {
		t.Fatal(err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	addrs, _ := hst.Addrs()
	_, err = hst2.Connect(ctx, hst.Config().ID, addrs)
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Println(addrs)
}

// 测试发送protobuf消息。注册消息处理器
func TestConnSendProtobufMsgAndHandler(t *testing.T) {
	var localMa = "/ip4/0.0.0.0/tcp/9040"
	var localMa2 = "/ip4/0.0.0.0/tcp/9042"

	ma, _ := multiaddr.NewMultiaddr(localMa)
	ma2, _ := multiaddr.NewMultiaddr(localMa2)

	// 创建节点1
	hst, _ := host.NewHost(option.ListenAddr(ma))
	t.Log(hst.Addrs())
	// 注册消息处理器
	err := hst.RegisterMsgHandler(0x11, func(msgId uint16, data []byte, handler *pbMsgHandler.PBMsgHandler) {
		var msg pb.StringMsg
		err := proto.Unmarshal(data, &msg)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(msg.Value)
	})
	// 开启服务监听消息
	go hst.Serve(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// 创建节点2
	hst2, yterr := host.NewHost(option.ListenAddr(ma2))
	if yterr != nil {
		t.Fatal(yterr.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_, err = hst2.Connect(ctx, hst.Config().ID, []multiaddr.Multiaddr{ma})
	if err != nil {
		t.Fatal(err)
	}
	var msg pb.StringMsg
	msg.Value = "测试buf消息"

	// 发送消息
	yerr := hst2.SendMsg("test", 0x11, &msg)
	if yerr != nil {
		t.Fatal(yerr.Error())
	}
	<-time.After(3 * time.Second)
}
