package host_test

import (
	host "YTHost"
	"YTHost/Tool"
	"encoding/hex"
	"fmt"
	ytcrypto "github.com/yottachain/YTCrypto"
	"testing"
)

var localMa = "/ip4/0.0.0.0/tcp/9000"
var localMa2 = "/ip4/0.0.0.0/tcp/9001"
var localMa3 = "/ip4/0.0.0.0/tcp/9002"

func TestMd5(t *testing.T)  {
	test := "test"
	fmt.Println(hex.EncodeToString(Tool.Md5(test)))
	fmt.Println(len(Tool.Md5(test)))
	fmt.Println(host.GetMethodSign(test))
	fmt.Println(len(host.GetMethodSign(test)))
}

func TestAddr(t *testing.T)  {
	privKey, _ := ytcrypto.CreateKey()
	h := host.NewHost(privKey, localMa)
	fmt.Println(h.Addrs())
}

func TestNewHost(t *testing.T) {
	privKey, _ := ytcrypto.CreateKey()
	h := host.NewHost(privKey, localMa2)
	fmt.Println(h.Addrs())
}

// TestSendMessage 测试发送接受消息
func TestStartServer(t *testing.T) {
	fmt.Println("test start server.....")
	// 创建host1模拟接受消息
	privKey, _ := ytcrypto.CreateKey()
	h1 := host.NewHost(localMa, privKey)
	fmt.Println("new host .....")

	//tcpListener := h1.NewListener("127.0.0.1", "8980")
	tcpListener := h1.NewListener("0.0.0.0", "8980")
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
	privKey, _ := ytcrypto.CreateKey()

	h2 := host.NewHost(privKey, localMa2)

	//// 连接节点1
	addrs := make([]string, 0)
	addrs = append(addrs, "127.0.0.1:8980")
	//addrs = append(addrs, "172.20.10.2:8980")
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

func TestHst_ID(t *testing.T) {
	privKey, _ := ytcrypto.CreateKey()
	h := host.NewHost(privKey, localMa2)
	fmt.Println(len(h.ID()))
}

