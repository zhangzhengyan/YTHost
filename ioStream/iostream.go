package ioStream

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"
)

const (
	RES = 'c'
	RPS = 's'

)

var testCount  = 0

func NewStreamHandler(conn io.ReadWriteCloser) (sconn *ReadWriteCloser, cconn *ReadWriteCloser){
	l := sync.Mutex{}

	s := NewReadWriter()
	sconn = &s
	c := NewReadWriter()
	cconn = &c

	buf := bufio.NewWriter(conn)

	//num := testCount + 1
	testCount ++

	go func() {
		by := make([]byte, 2)
		for {
			if sconn.isClose == true || cconn.isClose == true {
				sconn.SetReadErr()
				cconn.SetReadErr()
				return
			}
			f, _, msg, err := DecodeConn(conn, by)
			///time.Sleep(time.Second*2)
			//fmt.Printf("count:%d----f:%s\n", num, string(f))
			//fmt.Printf("count:%d" + "----msg:" + string(msg) + "\n", testCount)
			if err != nil {
				fmt.Println(err)
				_ = sconn.Close()
				_ = cconn.Close()
			}
			if f == RES {
				_ = sconn.ReadAppend(msg)
			}else if f == RPS {
				_ = cconn.ReadAppend(msg)
			}else {
				continue
			}
		}
	}()

	go func(l *sync.Mutex) {
		msg := make([]byte, 2048 + 3)
		for {
			if sconn.isClose == true {
				return
			}
			n, err:= sconn.WriteConsume(2048, RPS, msg)
			//fmt.Printf("sconn----->>>>>count:%d----n:%d\n", num, n)

			if err != nil {
				fmt.Println(err)
			}
			if n > 0 {
				l.Lock()
				//fmt.Println(msg[0:n+3])
				buf.Write(msg[0:n+3])
				buf.Flush()
				l.Unlock()
			}
		}
	}(&l)

	go func(l *sync.Mutex) {
		msg := make([]byte, 2048 + 3)
		for {
			if cconn.isClose == true {
				return
			}
			n, err:= cconn.WriteConsume(2048, RES, msg)
			//fmt.Printf("cconn----->>>>>count:%d----n:%d\n",  num, n)

			if err != nil {
				fmt.Println(err)
			}
			if n > 0 {
				l.Lock()
				//fmt.Println(msg[0:n+3])
				buf.Write(msg[0:n+3])
				buf.Flush()
				l.Unlock()
			}
		}
	}(&l)

	return
}

func byteToInt16(b []byte) (uint16, error) {
	if len(b) > 2 {
		return 0, errors.New("bytes lenth must less 3\n")
	}
	bytebuff := bytes.NewBuffer(b)
	var ret uint16
	err := binary.Read(bytebuff, binary.BigEndian, &ret)
	if err != nil {
		return 0, err
	}

	return ret, nil
}

func Int16Tobyte(m uint16) ([]byte, error) {

	buf := bytes.NewBuffer([]byte{})
	err := binary.Write(buf, binary.BigEndian, m)

	return buf.Bytes(), err
}

func DecodeConn(conn io.ReadWriteCloser, buf []byte) (flag byte, bLen uint16, msgbuf []byte, err error) {
	//第一个字节是标志 后两个字节是长度
	n, err := io.ReadFull(conn, buf[0:1])
	if n == 0 {
		return
	}
	b := buf[0]
	if b != RES && b != RPS {
		err = errors.New("conn stream flag error\n")
	}
	flag = b

	//两个字节作为每个消息块的长度
	n, err = io.ReadFull(conn, buf[0:2])
	if n == 0 {
		return
	}

	bLen, err = byteToInt16(buf)
	if err == nil {
		msgbuf = make([]byte, bLen)
		n, err = io.ReadFull(conn, msgbuf)
		bLen = uint16(n)
	}

	return
}

type Reader struct {
	buf 		[]byte
	r, w       	int
	l   		sync.Mutex
}

const (
	ERR_BUFNOZERO = "read buf lenth can't be zero\n"
	ERR_BUFOVERFLOW = "read buf overflow\n"
	INT_MAX = int(^uint(0) >> 1)
)

func (b * Reader) Read(p [] byte) (n int, err error){
	n = len(p)
	if n == 0 {
		err = errors.New(ERR_BUFNOZERO)
		return
	}

	b.l.Lock()
	defer b.l.Unlock()

	if b.r < 0 {
		err = io.EOF
		return
	}

	// copy as much as we can
	n = copy(p, b.buf[b.r:b.w])
	b.r += n
	b.buf = b.buf[b.r:b.w]
	b.w = b.w - b.r
	b.r = 0

	return n, nil
}

func (b *Reader) SetReadErr(){
	b.l.Lock()
	defer b.l.Unlock()
	b.r = -1
}

//app data to buf
func (b * Reader) ReadAppend(p [] byte) (err error){
	b.l.Lock()
	defer b.l.Unlock()
	lenth := len(p)
	b.w = b.w + lenth
	if b.w > INT_MAX {
		err = errors.New(ERR_BUFOVERFLOW)
	}else {
		b.buf = append(b.buf, p...)
	}

	return
}

func NewReader() *Reader {
	r := new(Reader)
	return r
}

type Writer struct {
	buf 		[]byte
	n       	int
	l   		sync.Mutex
	wc   		chan bool
}

const (
	defaultBufSize = 4096
)

func (b *Writer) Available() int { return len(b.buf) - b.n }

func (b *Writer) Buffered() int { return b.n }

func (b *Writer) Write(p []byte) (nn int, err error) {
	for  {
		b.l.Lock()
		if len(p) > b.Available() {
			n := copy(b.buf[b.n:], p)
			p = p[n:]
			b.n += n
			nn += n
		}else {
			n := copy(b.buf[b.n:], p)
			b.n += n
			nn += n
			b.l.Unlock()
			b.wc <- true
			break
		}
		b.l.Unlock()
		b.wc <- true
	}

	return nn, nil
}

func (b *Writer) WriteConsume(n int, flag byte, msg []byte) ( nn int, err error){
	if n <= 0 {
		return
	}

	<- b.wc

	b.l.Lock()
	defer b.l.Unlock()


	if b.Buffered() <= 0 {
		nn = 0
		return
	}

	//前三个字节作为标识
	//msg = make([]byte, n + 3)
	msg[0] = flag
	nn = copy(msg[3:], b.buf[:b.n])
	b.n = b.n - nn
	//b.buf = b.buf[nn:]
	copy(b.buf, b.buf[nn:])
	mLenbyte, err := Int16Tobyte(uint16(nn))
	msg[1] = mLenbyte[0]
	msg[2] = mLenbyte[1]

	return
}

func NewWriterSize(size int) *Writer {
	if size <= 0 {
		size = defaultBufSize
	}
	return &Writer{
		buf: make([]byte, size),
		n:  0,
	}
}

// NewWriter returns a new Writer whose buffer has the default size.
func NewWriter() *Writer {
	return NewWriterSize(defaultBufSize)
}

type Closer struct {
	isClose 	bool
}

func (c * Closer) Close() error{
	c.isClose = true
	return nil
}

type ReadWriteCloser struct {
	Reader
	Writer
	Closer
}

func NewReadWriter() ReadWriteCloser {
	r := Reader{
		buf: make([]byte, 1),
		r:   0,
		w:   0,
		l:   sync.Mutex{},
	}
	w := Writer{
		buf: make([]byte, defaultBufSize),
		n:   0,
		l:   sync.Mutex{},
		wc:  make(chan bool),
	}
	c := Closer{false}
	return ReadWriteCloser{r, w, c}
}