package service_test

import (
	host "github.com/graydream/YTHost"
	"github.com/graydream/YTHost/option"
	"github.com/graydream/YTHost/service"
	"github.com/multiformats/go-multiaddr"
	"golang.org/x/net/context"
	"testing"
	"time"
)


func TestMsg(t *testing.T){
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
		return []byte("msg test!!!!!!!!!!!"), nil
	})

	hst.RegisterHandler(0x11, func(requestData []byte, head service.Head) (bytes []byte, e error) {
		t.Log(string(requestData), head.RemotePubKey, head.RemotePeerID, head.RemoteAddrs)
		return []byte("msg id 0x11 test succeed"), nil
	})

	go hst.Accept()

	//---------------发送信息处理
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
	clt, err := hst1.Connect(ctx, hst.Config().ID, hst.Addrs())
	// 关闭连接
	defer clt.Close()
	if err != nil {
		t.Fatal(err.Error())
	}

	if res, err := clt.SendMsg(context.Background(), 0x11, []byte("111111111111111")); err != nil {
		t.Fatal(err)
	} else {
		t.Log(string(res))
	}

	if res, err := clt.SendMsg(context.Background(), 0x0, []byte("global msg")); err != nil {
		t.Fatal(err)
	} else {
		t.Log(string(res))
	}

	hst.RemoveGlobalHandler()
	hst.RemoveHandler(0x0)
}
