package host

import (
	"YTHost/Tool"
	"encoding/hex"
	"fmt"
	"github.com/multiformats/go-multiaddr"
	"net"
	"time"
	"yamux"
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
	NewListener(listenAddr string, port string) *net.TCPListener
	StartServer(tcplistener *net.TCPListener)
}

type hst struct {
	// TODO
	serverSession *yamux.Session // server端的session
	outConnects map[string][]net.Conn // peerID - []Conn
	outSessions map[string][]*yamux.Session // peerID - []Session
	methodHandlerMap map[string]MsgHandlerFunc // 方法名-具体方法
	methodSignMap map[string]string // 方法签名-方法名
	peeridmethodStreamMap map[string]net.Conn // peerID+方法名 - 流
	addrs []string // 本地内网、外网地址 + 对应端口
}

func NewHost(privateKey string, listenAddrs ...string) Host {
	// 获取本机端口
	ports := Tool.GetPortsFromMultiAddr(listenAddrs)
	if len(ports) <= 0 {
		return nil
	}
	// 拼装addr
	multiAddr := Tool.GetMultiAddr(listenAddrs)
	return &hst{serverSession:nil,
		outConnects:make(map[string][]net.Conn),
		outSessions:make(map[string][]*yamux.Session),
		methodHandlerMap:make(map[string]MsgHandlerFunc),
		methodSignMap:make(map[string]string),
		peeridmethodStreamMap:make(map[string]net.Conn),
		addrs:multiAddr,
	}
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
	return h.addrs
}

// multiAddr 转 常规地址

// 常规地址 转 multiAddr
func StringListToMaddrs(addrs []string) ([]multiaddr.Multiaddr, error) {
	maddrs := make([]multiaddr.Multiaddr, len(addrs))
	for k, addr := range addrs {
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return maddrs, err
		}
		maddrs[k] = maddr
	}
	return maddrs, nil
}

// Connect 连接节点
func (h *hst) Connect(id string, addrs []string) error {
	// map[string][]*yamux.Session
	conns := make([]net.Conn, 0)
	idConns := make(map[string][]net.Conn)
	sessions := make([]*yamux.Session, 0)
	idSessions := make(map[string][]*yamux.Session)
	for _, addr := range addrs {
		//conn, err := net.Dial("tcp", addr)
		// TODO 连接时间 设定需要进一步确认。
		conn, err := net.DialTimeout("tcp", addr, 10000 * time.Second)
		if err != nil {
			continue
		}
		conns = append(conns, conn)
		session, _ := yamux.Client(conn, nil)
		sessions = append(sessions, session)
	}
	idSessions[id] = sessions
	idConns[id] = conns
	h.outSessions = idSessions
	h.outConnects = idConns
	return nil
}

// DisConnect 断开连接
func (h *hst) DisConnect(id string) error {
	delete(h.outSessions, id)
	delete(h.outConnects, id)
	return nil
}

func (h *hst) SendMsg(id string, msgType string, msg []byte) ([]byte, error) {
	var stream net.Conn
	// 选择流，如果没有，则创建一个，然后存入到缓存中。
	if _, ok := h.peeridmethodStreamMap[id + msgType]; !ok {
		// 选择session
		if _, okSession := h.outSessions[id]; !okSession {
			// session 不存在，让用户先去Connect
			return []byte{}, Tool.YTError("method's session not connect~")
		}
		// 如果一个id 有多个session，暂时获取第一个
		stream, _ = h.outSessions[id][0].Open()
		// 将当前的流 存入缓存，供后续使用
		// TODO 下面这一句要加上。
		h.peeridmethodStreamMap[id + msgType] = stream
	} else {
		stream = h.peeridmethodStreamMap[id + msgType]
	}
	// 封装数据，将方法前面加入到数据流前面，便于在接收端处理。
	sendData := EncodeData(msgType, msg)
	// 发送数据
	stream.Write(sendData)
	// 等待响应
	// TODO 需要验证如果流复用，后面的数据是否能够拿到，并且需要验证准确度。
	replyBuf := make([]byte, 1024)
	stream.Read(replyBuf)
	return replyBuf, nil
}

// 封装数据
func EncodeData(msgType string, msg []byte) []byte {
	mtBytes := Tool.Md5(msgType)
	fmt.Println("send msgtype is ", hex.EncodeToString(mtBytes))
	return Tool.BytesCombine(mtBytes, msg)
}

// RegisterHandler 注册消息回调函数
func (h *hst) RegisterHandler(msgType string, MessageHandler MsgHandlerFunc) {
	// 方法注册
	if _, ok := h.methodHandlerMap[msgType]; !ok {
		// 注册 方法签名-方法名，便于流数据解析。
		fmt.Println("register method sign key:", GetMethodSign(msgType))
		h.methodSignMap[GetMethodSign(msgType)] = msgType
		// 注册   方法-handler
		fmt.Println("register method handler key:", msgType)
		h.methodHandlerMap[msgType] = MessageHandler
	}
}

// 返回特定长度的方法签名，便后后续解析
func GetMethodSign(msgType string) string  {
	return Tool.Md5String(msgType)
}

// unregisterHandler 移除消息处理器
func (h *hst) UnregisterHandler(msgType string) {
	if _, ok := h.methodHandlerMap[msgType]; ok {
		delete(h.methodHandlerMap, msgType)
	}
}

// Close 关闭
func (h *hst) Close() {
	// 关闭节点时，把所有连接都关掉
	outConns := h.outConnects
	if len(outConns) > 0 {
		for id, connlist := range outConns {
			if len(connlist) > 0 {
				for _, conn := range connlist {
					conn.Close()
				}
			}
			delete(outConns, id)
		}
	}
}

func (h *hst)NewListener(listenAddr string, port string) *net.TCPListener{
	tcpaddr, _ := net.ResolveTCPAddr("tcp4", listenAddr + ":" + port)
	tcplisten, _ := net.ListenTCP("tcp", tcpaddr)
	return tcplisten
}

func (h *hst) StartServer(tcplisten *net.TCPListener)  {
	conn, _ := tcplisten.Accept()
	session, _ := yamux.Server(conn, nil)
	h.serverSession = session

	fmt.Println("start server")
	id := 1
	for {
		// 建立多个流通路
		stream, err := session.Accept()
		if err == nil {
			// 实际的处理
			go h.Recv(stream, id)
		}else{
			fmt.Println("session over.")
			return
		}
	}
}

func (h *hst) Recv(stream net.Conn, id int){

	for {
		// TODO 这儿需要确认一下最大接收长度。 暂定1024
		// 接收请求
		buf := make([]byte, 1024)
		stream.Read(buf)
		// 获取消息类型
		msgTypeSign := hex.EncodeToString(buf[:16])
		fmt.Println("receive msgtype sign is ", msgTypeSign)
		var msgType string
		if _, ok := h.methodSignMap[msgTypeSign]; !ok {
			fmt.Println("msg sign not exist!!!")
		} else {
			msgType = h.methodSignMap[msgTypeSign]
		}
		// 实际请求
		if handler, ok := h.methodHandlerMap[msgType]; ok {
			msgData := buf[16:]
			// 处理
			reply, err := handler(msgType, msgData, "")
			fmt.Println("reply id:", id, " reply buf is ", reply)
			if err == nil {
				// 响应结果
				stream.Write(reply)
			}
		} else {
			fmt.Println("method not found~~~~")
		}
	}
}


