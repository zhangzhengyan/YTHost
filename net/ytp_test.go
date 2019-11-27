package net

import (
	"bytes"
	"context"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
	"testing"
	"time"
)

func TestHandshake(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	ytp := YTP{ReadWriter: buf, LocalID: "16Uiu2HAkuTUCLMEvN4UeCSsYLPTbdcWLoyMtjPhY9DaQ6CdCUwoU", LocaAddrs: []multiaddr.Multiaddr{}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	go func() {
		if err := ytp.Handshake(ctx); err != nil {
			t.Fatal(err)
		} else {
			t.Log(ytp.RemoteID, ytp.RemoteAddrs)
		}
	}()
	buf.WriteString("YTHost: 0.0.1\nID:16Uiu2HAkuTUCLMEvN4UeCSsYLPTbdcWLoyMtjPhY9DaQ6CdCUwoU\nAddr:/ip4/127.0.0.1/tcp/9001\n|end|")
	<-time.After(time.Second)
}

func TestNetAddress(t *testing.T) {
	addrs, err := manet.InterfaceMultiaddrs()
	if err != nil {
		t.Fatal(err)
	}
	for _, addr := range addrs {
		t.Log(addr.String())
	}
	m1, _ := multiaddr.NewMultiaddr("/ip4/0.0.0.0")
	m2, _ := multiaddr.NewMultiaddr("/tcp/9001")
	m1.Encapsulate(m2)
	t.Log()
}
