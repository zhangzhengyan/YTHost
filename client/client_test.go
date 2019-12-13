package client_test

import (
	host "github.com/graydream/YTHost"
	"github.com/graydream/YTHost/option"
	"github.com/graydream/YTHost/service"
	"github.com/multiformats/go-multiaddr"
	"golang.org/x/net/context"
	"testing"
	"time"
)

func TestClient(t *testing.T){
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

	ma, err = multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/9099")
	if err != nil {
		t.Log(err.Error())
	}
	hst1, err := host.NewHost(option.ListenAddr(ma))
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	clt, err := hst1.Connect(ctx, hst.Config().ID, hst.Addrs())
	// 关闭连接
	defer clt.Close()
	if err != nil {
		t.Fatal(err.Error())
	}

	lpInfo := clt.LocalPeer()
	t.Log(lpInfo)

	rpInfo := clt.RemotePeer()
	t.Log(rpInfo)

	pk := clt.RemotePeerPubkey()
	t.Log(pk)

	if res, err := clt.SendMsg(context.Background(), 0x0, []byte("2222")); err != nil {
		t.Fatal(err)
	} else {
		t.Log(string(res))
	}

	if res, err := clt.SendMsgClose(context.Background(), 0x0, []byte("2222")); err != nil {
		t.Fatal(err)
	} else {
		t.Log(string(res))
	}

	if !clt.IsClosed() {
		clt.Close()
	}
}
