package msgredis

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

const (
	ConnectTimeout     = 5e9
	ReadTimeout        = 5e9
	WriteTimeout       = 5e9
	DefaultBufferSize  = 64
	DefaultArgsBufSize = 10

	TypeError        = '-'
	TypeSimpleString = '+'
	TypeBulkString   = '$'
	TypeIntegers     = ':'
	TypeArrays       = '*'
)

var (
	ErrNil = errors.New("nil data return")
)

type Conn struct {
	pipeCount      int
	multiCount     int
	conn           *net.TCPConn
	lastActiveTime int64
	keepAlive      bool
	buffer         []byte
	argsBuf        []interface{}
	rb             *bufio.Reader
	wb             *bufio.Writer
	readTimeout    time.Duration
	writeTimeout   time.Duration
}

func NewConn(conn *net.TCPConn, connectTimeout, readTimeout, writeTimeout time.Duration, keepAlive bool) *Conn {
	return &Conn{
		conn:           conn,
		lastActiveTime: time.Now().Unix(),
		keepAlive:      keepAlive,
		buffer:         make([]byte, DefaultBufferSize),
		argsBuf:        make([]interface{}, DefaultArgsBufSize),
		rb:             bufio.NewReader(conn),
		wb:             bufio.NewWriter(conn),
		// rb:           bufio.NewReaderSize(conn, 16),
		// wb:           bufio.NewWriterSize(conn, 16),
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
	}
}

// connect with timeout
func Dial(address, password string, connectTimeout, readTimeout, writeTimeout time.Duration, keepAlive bool) (*Conn, error) {
	c, e := net.DialTimeout("tcp", address, connectTimeout)
	if e != nil {
		return nil, e
	}
	if _, ok := c.(*net.TCPConn); !ok {
		return nil, errors.New("invalid tcp conn")
	}

	conn := NewConn(c.(*net.TCPConn), connectTimeout, readTimeout, writeTimeout, keepAlive)
	if password != "" {
		if _, e := conn.AUTH(password); e != nil {
			return nil, e
		}
	}
	return conn, nil
}

func (c *Conn) Close() {
	if c.conn != nil {
		c.Close()
	}
}

// call redis command with request => response model
func (c *Conn) Call(command string, args ...interface{}) (interface{}, error) {
	c.lastActiveTime = time.Now().Unix()
	var e error
	if c.writeTimeout > 0 {
		if e = c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout)); e != nil {
			return nil, e
		}
	}
	if e = c.writeRequest(command, args); e != nil {
		return nil, e
	}

	if e = c.wb.Flush(); e != nil {
		return nil, e
	}

	if c.readTimeout > 0 {
		if e = c.conn.SetReadDeadline(time.Now().Add(c.writeTimeout)); e != nil {
			return nil, e
		}
	}
	response, e := c.readResponse()
	if e != nil {
		return nil, e
	}
	return response, e
}

// write response
func (c *Conn) writeRequest(command string, args []interface{}) error {
	fmt.Println("REQUEST:", command, args)
	var e error
	if e = c.writeLen('*', 1+len(args)); e != nil {
		return e
	}

	if e = c.writeString(command); e != nil {
		return e
	}

	for _, arg := range args {
		if e != nil {
			return e
		}
		switch data := arg.(type) {
		case int:
			e = c.writeInt64(int64(data))
		case int64:
			e = c.writeInt64(data)
		case float64:
			e = c.writeFloat64(data)
		case string:
			e = c.writeString(data)
		case []byte:
			e = c.writeBytes(data)
		case bool:
			if data {
				e = c.writeString("1")
			} else {
				e = c.writeString("0")
			}
		case nil:
			e = c.writeString("")
		default:
			e = c.writeString(fmt.Sprintf("%v", data))
		}
	}
	return e
}

// reuse one buffer
func (c *Conn) writeLen(prefix byte, n int) error {
	pos := len(c.buffer) - 1
	c.buffer[pos] = '\n'
	pos--
	c.buffer[pos] = '\r'
	pos--

	for i := n; i != 0 && pos >= 0; i = i / 10 {
		c.buffer[pos] = byte(i%10 + '0')
		pos--
	}
	c.buffer[pos] = prefix
	_, e := c.wb.Write(c.buffer[pos:])
	if e != nil {
		return e
	}
	return nil
}

// write
func (c *Conn) writeBytes(b []byte) error {
	var e error
	if e = c.writeLen('$', len(b)); e != nil {
		return e
	}
	if _, e = c.wb.Write(b); e != nil {
		return e
	}
	if _, e = c.wb.WriteString("\r\n"); e != nil {
		return e
	}
	return nil
}

func (c *Conn) writeString(s string) error {
	var e error
	if e = c.writeLen('$', len(s)); e != nil {
		return e
	}
	if _, e = c.wb.WriteString(s); e != nil {
		return e
	}
	if _, e = c.wb.WriteString("\r\n"); e != nil {
		return e
	}
	return nil

}

func (c *Conn) writeFloat64(f float64) error {
	// Negative precision means "only as much as needed to be exact."
	return c.writeBytes(strconv.AppendFloat([]byte{}, f, 'g', -1, 64))
}

func (c *Conn) writeInt64(n int64) error {
	return c.writeBytes(strconv.AppendInt([]byte{}, n, 10))
}

// read
func (c *Conn) readResponse() (interface{}, error) {
	var e error
	p, e := c.readLine()
	if e != nil {
		return nil, e
	}
	resType := p[0]
	p = p[1:]
	switch resType {
	case TypeError:
		return nil, errors.New(string(p))
	case TypeIntegers:
		return strconv.ParseInt(string(p), 10, 64)
	case TypeSimpleString:
		return p, nil
	case TypeBulkString:
		return c.parseBulkString(p)
	case TypeArrays:
		return c.parseArray(p)
	default:
	}

	return nil, errors.New("illegal response type")
}

func (c *Conn) readLine() ([]byte, error) {
	var e error
	p, e := c.rb.ReadBytes('\n')
	if e != nil {
		return nil, e
	}

	i := len(p) - 2
	if i <= 0 {
		return nil, errors.New("invalid terminator")
	}
	return p[:i], nil
}

// func (c *Conn) parseInt(p []byte) (int64, error) {
// 	n, e := strconv.ParseInt(string(p), base, bitSize)
// 	if e != nil {
// 		return 0, e
// 	}
// 	return n, nil
// }

func (c *Conn) parseBulkString(p []byte) ([]byte, error) {
	n, e := strconv.ParseInt(string(p), 10, 64)
	if e != nil {
		return []byte{}, e
	}
	if n == -1 {
		return nil, nil
	}

	result := make([]byte, n+2)
	_, e = io.ReadFull(c.rb, result)
	return result[:n], e
}

func (c *Conn) parseArray(p []byte) ([]interface{}, error) {
	n, e := strconv.ParseInt(string(p), 10, 64)
	if e != nil {
		return nil, e
	}

	if n == -1 {
		return nil, nil
	}

	result := make([]interface{}, n)
	var i int64
	for ; i < n; i++ {
		result[i], e = c.readResponse()
		if e != nil {
			return nil, e
		}
	}
	return result, nil
}

// pipeline
func (c *Conn) PipeSend(command string, args ...interface{}) error {
	c.pipeCount++
	return c.writeRequest(command, args)
}

func (c *Conn) PipeExec() ([]interface{}, error) {
	var e error
	if e = c.wb.Flush(); e != nil {
		return nil, e
	}
	n := c.pipeCount
	ret := make([]interface{}, c.pipeCount)
	c.pipeCount = 0
	for i := 0; i < n; i++ {
		ret[i], e = c.readResponse()
	}
	return ret, e
}

// Transactions
func (c *Conn) MULTI() error {
	ret, e := c.Call("MULTI")
	if e != nil {
		return e
	}
	if _, ok := ret.([]byte); !ok {
		return errors.New("invalid return type")
	}
	r := ret.([]byte)
	if len(r) == 2 && r[0] == 'O' && r[1] == 'K' {
		return nil
	}
	return errors.New("invalid return:" + string(r))
}

func (c *Conn) TransSend(command string, args ...interface{}) error {
	c.multiCount++
	ret, e := c.Call(command, args...)
	if e != nil {
		return e
	}
	if _, ok := ret.([]byte); !ok {
		return errors.New("invalid return type")
	}
	r := ret.([]byte)
	if len(r) == 2 && r[0] == 'O' && r[1] == 'K' {
		return nil
	}
	return errors.New("invalid return:" + string(r))
}

func (c *Conn) TransExec() ([]interface{}, error) {
	r, e := c.Call("EXEC")
	if e = c.wb.Flush(); e != nil {
		return nil, e
	}
	return r.([]interface{}), e
}
