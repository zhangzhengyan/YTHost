package host_test

import (
	"context"
	"fmt"
	"github.com/gogo/protobuf/proto"
	host "github.com/graydream/YTHost"
	"github.com/graydream/YTHost/DataFrameEncoder"
	"github.com/graydream/YTHost/option"
	"github.com/graydream/YTHost/pb"
	"github.com/graydream/YTHost/pbMsgHandler"
	"github.com/multiformats/go-multiaddr"
	"io"
	"testing"
	"time"
)

// 测试创建通讯节点
func TestNewHost(t *testing.T) {
	var localMa2 = "/ip4/0.0.0.0/tcp/9001"

	ma, _ := multiaddr.NewMultiaddr(localMa2)
	if hst, err := host.NewHost(option.ListenAddr(ma)); err != nil {
		t.Fatal(err)
	} else {
		t.Log(hst.Listenner().Addr())
	}
}

// 测试建立连接
func TestConn(t *testing.T) {
	var localMa = "/ip4/0.0.0.0/tcp/9010"
	var localMa2 = "/ip4/0.0.0.0/tcp/9011"

	ma, _ := multiaddr.NewMultiaddr(localMa)
	ma2, _ := multiaddr.NewMultiaddr(localMa2)
	hst, _ := host.NewHost(option.ListenAddr(ma))
	t.Log(hst.Listenner().Addr())
	go func() {
		for {
			conn, err := hst.Listenner().Accept()
			if err != nil {
				t.Fatal(err)
			} else if conn.RemoteAddr() != nil {
				t.Log(conn.RemoteAddr())
			}
		}
	}()
	hst2, _ := host.NewHost(option.ListenAddr(ma2))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := hst2.Connect(ctx, "id111", []multiaddr.Multiaddr{ma})
	if err != nil {
		t.Fatal(err.Msg)
	} else if conn.RemoteAddr() != nil {
		t.Log(conn.RemoteAddr())
	}
	conn, err = hst2.Connect(ctx, "id111", []multiaddr.Multiaddr{ma})
	if err != nil {
		t.Fatal(err.Msg)
	} else if conn.RemoteAddr() != nil {
		t.Log(conn.RemoteAddr())
	}
}

// 测试发送消息
func TestConnSendMsg(t *testing.T) {
	var localMa = "/ip4/0.0.0.0/tcp/9020"
	var localMa2 = "/ip4/0.0.0.0/tcp/9021"

	ma, _ := multiaddr.NewMultiaddr(localMa)
	ma2, _ := multiaddr.NewMultiaddr(localMa2)
	hst, _ := host.NewHost(option.ListenAddr(ma))
	t.Log(hst.Listenner().Addr())
	go func() {
		for {
			conn, err := hst.Listenner().Accept()
			if err != nil {
				t.Fatal(err)
			} else {
				t.Log(conn.RemoteAddr())
				md := dataFrameEncoder.NewDecoder(conn)
				for {
					msg, _ := md.Decode()
					fmt.Println(string(msg))
				}
			}
		}
	}()
	hst2, _ := host.NewHost(option.ListenAddr(ma2))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := hst2.Connect(ctx, "id111", []multiaddr.Multiaddr{ma})
	if err != nil {
		t.Fatal(err.Msg)
	} else {
		t.Log(conn.RemoteAddr())
		me := dataFrameEncoder.NewEncoder(conn)
		me.Encode([]byte("测试数据"))
		<-time.After(time.Second)
		me.Encode([]byte("测试数据2"))
	}
}

// 测试发送protobuf消息
func TestConnSendProtobufMsg(t *testing.T) {
	var localMa = "/ip4/0.0.0.0/tcp/9030"
	var localMa2 = "/ip4/0.0.0.0/tcp/9031"

	ma, _ := multiaddr.NewMultiaddr(localMa)
	ma2, _ := multiaddr.NewMultiaddr(localMa2)
	hst, _ := host.NewHost(option.ListenAddr(ma))
	t.Log(hst.Listenner().Addr())
	go func() {
		for {
			conn, err := hst.Listenner().Accept()
			if err != nil {
				t.Fatal(err)
			} else {
				t.Log(conn.RemoteAddr())
				md := dataFrameEncoder.NewDecoder(conn)
				for {
					msgData, err := md.Decode()
					if err != nil && err.Error() == io.EOF.Error() {
						//t.Log("连接已关闭",err)
					} else {
						var msg pb.StringMsg
						if err := proto.Unmarshal(msgData, &msg); err != nil {
							t.Fatal(err)
						} else {
							fmt.Println(msg.Value)
						}
					}

				}
			}
		}
	}()
	hst2, _ := host.NewHost(option.ListenAddr(ma2))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := hst2.Connect(ctx, "id111", []multiaddr.Multiaddr{ma})
	if err != nil {
		t.Fatal(err.Msg)
	} else {
		t.Log(conn.RemoteAddr())
		me := dataFrameEncoder.NewEncoder(conn)
		var msg pb.StringMsg
		msg.Value = "测试protobuf消息"
		if msgData, err := proto.Marshal(&msg); err != nil {
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

// 测试发送protobuf消息。注册消息处理器
func TestConnSendProtobufMsgAndHandler(t *testing.T) {
	var localMa = "/ip4/0.0.0.0/tcp/9040"
	var localMa2 = "/ip4/0.0.0.0/tcp/9041"

	ma, _ := multiaddr.NewMultiaddr(localMa)
	ma2, _ := multiaddr.NewMultiaddr(localMa2)

	// 创建节点1
	hst, _ := host.NewHost(option.ListenAddr(ma))
	t.Log(hst.Listenner().Addr())
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
	hst2, _ := host.NewHost(option.ListenAddr(ma2))
	_, err = hst2.Connect(context.Background(), "test", []multiaddr.Multiaddr{ma})
	if err != nil {
		t.Fatal(err)
	}
	var msg pb.StringMsg
	msg.Value = "测试buf消息"

	// 发送消息
	yerr := hst2.SendMsg("test", 0x11, &msg)
	if err != nil {
		t.Fatal(yerr.Error())
	}
	<-time.After(3 * time.Second)
}
