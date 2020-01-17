package ioStream

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
)

const (
	RES = 'c'
	RPS = 's'
	defaultBufSize = 8192
)

var MAGIC = [3]byte{'Y', 'T', 'A'}

var testCount  = 0

func NewStreamHandler(conn io.ReadWriteCloser) (sconn *ReadWriteCloser, cconn *ReadWriteCloser){
	l := sync.Mutex{}

	s := NewReadWriter()
	sconn = &s
	c := NewReadWriter()
	cconn = &c

	buf := bufio.NewWriter(conn)

	testCount ++
	go func() {
		connect := conn
		by := make([]byte, 16)
		for {
			if sconn.GetClose() == true || cconn.GetClose() == true {
				sconn.SetReadErr()
				cconn.SetReadErr()
				_ = connect.Close()
				return
			}
			f, _, msg, err := DecodeConn(conn, by)
			///time.Sleep(time.Second*2)
			//fmt.Printf("count:%d----f:%s\n", num, string(f))
			//fmt.Printf("count:%d" + "----msg:" + string(msg) + "\n", testCount)
			if err != nil  {
				if err == io.EOF {
					log.Println(err)
					fmt.Println(err)
					_ = sconn.Close()
					_ = cconn.Close()
				}
				continue
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

	var WCfunc = func(l *sync.Mutex, conn *ReadWriteCloser, flag byte) {
		msg := make([]byte, 2048 + 6)
		for {
			if conn.isClose == true {
				return
			}
			n, err:= conn.WriteConsume(2048, flag, msg)
			//fmt.Printf("sconn----->>>>>count:%d----n:%d\n", num, n)

			if err != nil {
				fmt.Println(err)
				continue
			}
			if n > 0 {
				l.Lock()
				//fmt.Println(msg[0:n+3])
				_, err = buf.Write(msg[0:n+6])
				if nil == err {
					_ = buf.Flush()
				}
				l.Unlock()
			}
		}
	}

	go WCfunc(&l, sconn, RPS)
	go WCfunc(&l, cconn, RES)

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
	//第一个字节是标志 后面三个字节数魔术字 最后两个字节是长度
	n, err := io.ReadFull(conn, buf[0:1])
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err = io.EOF
		}
		return
	}

	b := buf[0]
	if b != RES && b != RPS {
		err = errors.New("conn stream flag error\n")
		return
	}
	flag = b

	//继续三个字节是魔术字
	n, err = io.ReadFull(conn, buf[0:3])
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err = io.EOF
		}
		return
	}

	magic:
	if buf[0] != MAGIC[0] || buf[1] != MAGIC[1] || buf[2] != MAGIC[2] {
		if buf[0] == RES || buf[0] == RPS {
			flag = buf[0]

			buf[0] = buf[1]
			buf[1] = buf[2]
			n, err = io.ReadFull(conn, buf[2:3])
			if err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					err = io.EOF
				}
				return
			}
			goto magic
		}
		err = errors.New("conn stream magic error\n")
		return
	}

	//继续两个字节作为每个消息块的长度
	n, err = io.ReadFull(conn, buf[0:2])
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err = io.EOF
		}
		return
	}

	bLen, err = byteToInt16(buf[0:2])
	if err == nil {
		msgbuf = make([]byte, bLen)
		n, err = io.ReadFull(conn, msgbuf)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				err = io.EOF
			}
			return
		}
		bLen = uint16(n)
	}

	return
}

type Reader struct {
	buf 		[]byte
	r, w       	int
	l   		sync.Mutex
	rc 			chan bool
}

const (
	ERR_BUFNOZERO = "read buf lenth can't be zero\n"
	INT_MAX = int(^uint(0) >> 1)
)

func (b *Reader) Available() int { return len(b.buf) - b.w }

func (b *Reader) SetReadErr(){
	b.l.Lock()
	defer b.l.Unlock()
	b.r = -1
}

func (b * Reader) Read(p [] byte) (n int, err error){
	n = len(p)
	if n == 0 {
		err = errors.New(ERR_BUFNOZERO)
		return
	}

	<- b.rc

	b.l.Lock()
	defer b.l.Unlock()

	if b.r < 0 {
		err = io.EOF
		return
	}

	// copy as much as we can
	n = copy(p, b.buf[b.r:b.w])
	b.r += n
	if b.Available() <= 0 {
		copy(b.buf, b.buf[b.r:b.w])
		b.w = b.w - b.r
		b.r = 0
	}

	return n, nil
}

//app data to buf
func (b * Reader) ReadAppend(p [] byte) (err error){
	for {
		b.l.Lock()
		if len(p) > b.Available() {
			copy(b.buf, b.buf[b.r:b.w])
			b.w = b.w - b.r
			b.r = 0
		}
		n := copy(b.buf[b.w:], p)
		b.w += n
		if n == len(p) {
			b.l.Unlock()
			b.rc <- true
			return
		}else {
			p = p[n:]
			b.l.Unlock()
			b.rc <- true
		}
	}
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
	msg[1] = MAGIC[0]
	msg[2] = MAGIC[1]
	msg[3] = MAGIC[2]
	nn = copy(msg[6:], b.buf[:b.n])
	b.n = b.n - nn
	copy(b.buf, b.buf[nn:])
	mLenbyte, err := Int16Tobyte(uint16(nn))
	msg[4] = mLenbyte[0]
	msg[5] = mLenbyte[1]

	return
}

func NewWriterSize(size int) *Writer {
	if size <= 0 {
		size = defaultBufSize
	}
	return &Writer{
		buf: make([]byte, size),
		n:  0,
		l:   sync.Mutex{},
		wc:  make(chan bool, 128),
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

func (c * Closer) GetClose() bool {
	return c.isClose
}

type ReadWriteCloser struct {
	Reader
	Writer
	Closer
}

func NewReadWriter() ReadWriteCloser {
	r := Reader{
		buf: make([]byte, defaultBufSize),
		r:   0,
		w:   0,
		l:   sync.Mutex{},
		rc:	 make(chan bool, 128),
	}
	w := Writer{
		buf: make([]byte, defaultBufSize),
		n:   0,
		l:   sync.Mutex{},
		wc:  make(chan bool, 128),
	}
	c := Closer{false}
	return ReadWriteCloser{r, w, c}
}