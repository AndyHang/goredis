package msgredis

import (
	"testing"
)

// func TestConn(t *testing.T) {
// 	c, e := Dial("10.16.15.121:9731", ConnectTimeout, ReadTimeout, WriteTimeout, false)
// 	if e != nil {
// 		println(e.Error())
// 		return
// 	}
// 	defer c.conn.Close()
// 	command := []byte{'P', 'I', 'N', 'G', '\r', '\n'}
// 	n, e := c.conn.Write(command)
// 	if e != nil {
// 		println(e.Error())
// 		return
// 	}
// 	println(n)

// 	response := make([]byte, 10)
// 	n, e = c.conn.Read(response)
// 	if e != nil {
// 		println(e.Error())
// 		return
// 	}
// 	println(n, string(response))
// }

// func TestDiv(t *testing.T) {
// 	p := []byte{'a', 'b'}
// 	t.Error(p[1:0])
// }

func TestWrite(t *testing.T) {
	c, e := Dial("10.16.15.121:9731", ConnectTimeout, ReadTimeout, WriteTimeout, false)
	if e != nil {
		println(e.Error())
		return
	}
	defer c.conn.Close()

	c.Call("set", "zyh0922", "zyh0922")
	c.Call("get", "zyh0922")
	c.Call("del", "zyh0922")
	c.Call("sadd", "zyh0922", "aaaaaa", "bbbbbbb", "ccccccccc")
	c.Call("smembers", "zyh0922")
	c.Call("smembers", "sadd")
	c.Call("del", "zyh0922")
}
