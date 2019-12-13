package clientStore_test

import (
	host "github.com/graydream/YTHost"
	"github.com/graydream/YTHost/option"
	"github.com/graydream/YTHost/service"
	"github.com/multiformats/go-multiaddr"
	"golang.org/x/net/context"
	"testing"
	"time"
)

func TestClientStore(t *testing.T) {
	ma, err := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/9009")
	if err != nil {
		t.Log(err.Error())
	}
	hst, err := host.NewHost(option.ListenAddr(ma))

	if err != nil {
		panic(err)
	}

	hst.RegisterGlobalMsgHandler(func(requestData []byte, head service.Head) (bytes []byte, e error) {
		t.Log(string(requestData), head.RemotePubKey, head.RemotePeerID, head.RemoteAddrs)
		return []byte("11111111111"), nil
	})

	go hst.Accept()

	//------------------
	ma, err = multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/9099")
	if err != nil {
		t.Log(err.Error())
	}
	hst1, err := host.NewHost(option.ListenAddr(ma))
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, err = hst1.ClientStore().Get(ctx, hst.Config().ID, hst.Addrs())
	if err != nil {
		t.Fatal(err)
	}

	// 第二次获取相同连接，不应建立新的连接
	_, err = hst1.ClientStore().Get(ctx, hst.Config().ID, hst.Addrs())
	if err != nil {
		t.Fatal(err)
	}
	if hst1.ClientStore().Len() > 1 {
		t.Fatal("不应该建立新连接")
	}

	//地址字符串连接测试
	maddrs := hst.Addrs()
	addrs := make([]string, len(maddrs))
	for k, m := range maddrs {
		addrs[k] = m.String()
	}
	_, err = hst1.ClientStore().GetByAddrString(ctx,  hst.Config().ID.String(), addrs)
	if err != nil {
		t.Fatal(err)
	}

	// 另一种发送消息的方式
	if res, err := hst1.SendMsg(context.Background(), hst.Config().ID, 0x11, []byte("2222")); err != nil {
		t.Fatal(err)
	} else {
		t.Log(string(res))
	}
	if err := hst1.ClientStore().Close(hst.Config().ID); err != nil {
		t.Fatal(err)
	}

	//再通过地址字符串连接一下
	_, err = hst1.ClientStore().GetByAddrString(ctx,  hst.Config().ID.String(), addrs)
	if err != nil {
		t.Fatal(err)
	}

	if err := hst1.ClientStore().Close(hst.Config().ID); err != nil {
		t.Fatal(err)
	}
}
