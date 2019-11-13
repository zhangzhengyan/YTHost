package host_test

import (
	host "YTHost"
	"YTHost/Tool"
	"encoding/hex"
	"fmt"
	"testing"
)

func TestMd5(t *testing.T)  {
	test := "test"
	fmt.Println(hex.EncodeToString(Tool.Md5(test)))
	fmt.Println(len(Tool.Md5(test)))
	fmt.Println(host.GetMethodSign(test))
	fmt.Println(len(host.GetMethodSign(test)))
}

func TestFuncInArray(t *testing.T)  {
	a := make(map[string]host.MsgHandlerFunc)
	a["ping"] = func(msgType string, msg []byte, publicKey string) ([]byte, error) {
		if string(msgType) == "ping" {
			return []byte("pong"), nil
		} else {
			return []byte("error"), nil
		}
	}
	reply, _ := a["ping"]("ping", []byte{}, "")
	fmt.Println(string(reply))
}

// TestSendMessage 测试发送接受消息
func TestStartServer(t *testing.T) {
	fmt.Println("test start server.....")
	// 创建host1模拟接受消息
	h1 := host.NewHost()
	fmt.Println("new host .....")

	tcpListener := h1.NewListener("127.0.0.1", "8980")
	fmt.Println("new server done .....")

	msgHandler := func(msgType string, msg []byte, publicKey string) ([]byte, error) {
		if string(msgType) == "ping" {
			return []byte("pong"), nil
		} else {
			return []byte("error"), nil
		}
	}
	h1.RegisterHandler("ping", msgHandler)
	fmt.Println("after register .....")
	h1.StartServer(tcpListener)

}

func TestSendMsg(t *testing.T)  {
	// 创建host2模拟发送消息
	h2 := host.NewHost()

	//// 连接节点1
	addrs := make([]string, 0)
	addrs = append(addrs, "127.0.0.1:8980")
	err := h2.Connect("1", addrs)
	if err != nil {
		t.Fatalf("connect err :%s", err)
	} else {
		t.Log("connect success")
	}
	//// 发送ping
	res, err := h2.SendMsg("1", "ping", []byte("ping"))
	if err != nil {
		t.Fatalf("sendMsg err :%s", err)

	} else {
		t.Log("sendMsg success")
		if string(res) == "pong" {
			t.Log("res success")
		} else {
			t.Fatal(string(res))
		}
	}
}

