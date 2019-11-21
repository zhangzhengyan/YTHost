package host_test

import (
	"context"
	"fmt"
	"github.com/gogo/protobuf/proto"
	host "github.com/graydream/YTHost"
	"github.com/graydream/YTHost/DataFrameEncoder"
	"github.com/graydream/YTHost/option"
	"github.com/graydream/YTHost/pb"
	"github.com/multiformats/go-multiaddr"
	"io"
	"testing"
	"time"
)

var localMa = "/ip4/0.0.0.0/tcp/9000"
var localMa2 = "/ip4/0.0.0.0/tcp/9001"
var localMa3 = "/ip4/0.0.0.0/tcp/9002"

// 测试创建通讯节点
func TestNewHost(t *testing.T) {
	ma,_ := multiaddr.NewMultiaddr(localMa2)
	if hst,err:=host.NewHost(option.ListenAddr(ma));err != nil {
		t.Fatal(err)
	} else {
		t.Log(hst.Listenner().Addr())
	}
}

// 测试建立连接
func TestConn(t *testing.T){
	ma,_ := multiaddr.NewMultiaddr(localMa)
	ma2,_:=multiaddr.NewMultiaddr(localMa2)
	hst,_:=host.NewHost(option.ListenAddr(ma))
	t.Log(hst.Listenner().Addr())
	go func() {
		for {
			Conn,err:=hst.Listenner().Accept()
			if err!=nil {
				t.Fatal(err)
			} else {
				t.Log(Conn.RemoteAddr())
			}
		}
	}()
	hst2,_:=host.NewHost(option.ListenAddr(ma2))
	ctx,cancel:=context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn,err:=hst2.Connect(ctx,[]multiaddr.Multiaddr{ma})
	if err != nil {
		t.Fatal(err.Msg)
	} else {
		t.Log(conn.RemoteAddr())
	}
	conn,err = hst2.Connect(ctx,[]multiaddr.Multiaddr{ma})
	if err != nil {
		t.Fatal(err.Msg)
	} else {
		t.Log(conn.RemoteAddr())
	}
}

// 测试发送消息
func TestConnSendMsg(t *testing.T){
	ma,_ := multiaddr.NewMultiaddr(localMa)
	ma2,_:=multiaddr.NewMultiaddr(localMa2)
	hst,_:=host.NewHost(option.ListenAddr(ma))
	t.Log(hst.Listenner().Addr())
	go func() {
		for {
			conn,err:=hst.Listenner().Accept()
			if err!=nil {
				t.Fatal(err)
			} else {
				t.Log(conn.RemoteAddr())
				md := dataFrameEncoder.NewDecoder(conn)
				for {
					msg,_:=md.Decode()
					fmt.Println(string(msg))
				}
			}
		}
	}()
	hst2,_:=host.NewHost(option.ListenAddr(ma2))
	ctx,cancel:=context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn,err:=hst2.Connect(ctx,[]multiaddr.Multiaddr{ma})
	if err != nil {
		t.Fatal(err.Msg)
	} else {
		t.Log(conn.RemoteAddr())
		me:= dataFrameEncoder.NewEncoder(conn)
		me.Encode([]byte("测试数据"))
		<-time.After(time.Second)
		me.Encode([]byte("测试数据2"))
	}
}

// 测试发送protobuf消息
func TestConnSendProtobufMsg(t *testing.T){
	ma,_ := multiaddr.NewMultiaddr(localMa)
	ma2,_:=multiaddr.NewMultiaddr(localMa2)
	hst,_:=host.NewHost(option.ListenAddr(ma))
	t.Log(hst.Listenner().Addr())
	go func() {
		for {
			conn,err:=hst.Listenner().Accept()
			if err!=nil {
				t.Fatal(err)
			} else {
				t.Log(conn.RemoteAddr())
				md := dataFrameEncoder.NewDecoder(conn)
				for {
					msgData,err:=md.Decode()
					if err!=nil&&err.Error() == io.EOF.Error(){
						t.Log("连接已关闭")
					} else {
						var msg pb.StringMsg
						if err := proto.Unmarshal(msgData,&msg);err != nil {
							t.Fatal(err)
						} else {
							fmt.Println(msg.Value)
						}
					}

				}
			}
		}
	}()
	hst2,_:=host.NewHost(option.ListenAddr(ma2))
	ctx,cancel:=context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn,err:=hst2.Connect(ctx,[]multiaddr.Multiaddr{ma})
	if err != nil {
		t.Fatal(err.Msg)
	} else {
		t.Log(conn.RemoteAddr())
		me:= dataFrameEncoder.NewEncoder(conn)
		var msg pb.StringMsg
		msg.Value = "测试protobuf消息"
		if msgData,err:=proto.Marshal(&msg);err != nil {
			t.Fatal(err)
		} else {
			if err := me.Encode(msgData); err != nil {
				t.Fatal(err)
			}
			conn.Close()
			if err := me.Encode(msgData); err != nil {
				t.Log(err)
			} else {
				t.Fatal("连接应该关闭")
			}
		}
	}
}