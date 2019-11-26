package net

import (
	"bytes"
	"context"
	"github.com/multiformats/go-multiaddr"
	"testing"
	"time"
)

func TestHandshake(t *testing.T){
	buf := bytes.NewBuffer([]byte{})
	buf.WriteString("YTHost: 0.0.1\nID:16Uiu2HAkuTUCLMEvN4UeCSsYLPTbdcWLoyMtjPhY9DaQ6CdCUwoU\nAddr:/ip4/127.0.0.1/tcp/9001\n\n\n")
	ytp:= YTP{ReadWriter:buf,LocalID:"16Uiu2HAkuTUCLMEvN4UeCSsYLPTbdcWLoyMtjPhY9DaQ6CdCUwoU",LocaAddrs:[]multiaddr.Multiaddr{}}
	ctx,cancel:=context.WithTimeout(context.Background(),time.Second*10)
	defer cancel()
	if err:=ytp.Handshake(ctx);err!=nil{
		t.Fatal(err)
	} else {
		t.Log(ytp.RemoteID,ytp.RemoteAddrs)
	}
}
