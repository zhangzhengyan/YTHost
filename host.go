package host

import (

)

type MsgHandlerFunc func(msgType string, msg []byte, publicKey string) ([]byte, error)

type Host interface {
	ConfigCallback(host string, port int32)
	ID() string
	Addrs() []string
	// Start()
	Connect(id string, addrs []string) error
	DisConnect(id string) error
	SendMsg(id string, msgType string, msg []byte) ([]byte, error)
	RegisterHandler(msgType string, MassageHandler MsgHandlerFunc)
	UnregisterHandler(msgType string)
	Close()
}

type hst struct {
	// TODO
}

func init()  {
	// TODO 参数注册
}

func (h *hst) ConfigCallback(host string, port int32) {
	// TODO
}

func (h *hst) ID() string {
	// TODO
	return ""
}

func (h *hst) Addrs() []string {
	length := 16
	addrs := make([]string, length)
	// TODO
	return addrs
}

// Connect 连接节点
func (h *hst) Connect(id string, addrs []string) error {
	// TODO
	return nil
}

// DisConnect 断开连接
func (h *hst) DisConnect(id string) error {
	// TODO
	return nil
}

func (h *hst) SendMsg(id string, msgType string, msg []byte) ([]byte, error) {
	b := make([]byte, 16)
	// TODO
	return b, nil
}

// RegisterHandler 注册消息回调函数
func (h *hst) RegisterHandler(msgType string, MessageHandler MsgHandlerFunc) {
	// TODO
}

// unregisterHandler 移除消息处理器
func (h *hst) UnregisterHandler(msgType string) {
	// TODO
}

// Close 关闭
func (h *hst) Close() {
	// TODO
}