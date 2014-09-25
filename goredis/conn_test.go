package msgredis

import (
	"fmt"
	"testing"
)

func TestPoolNoAuth(t *testing.T) {
	p := NewPool("10.16.15.121:9731", "")
	connSlice := make([]*Conn, 0, 25)
	for i := 0; i < 5; i++ {
		fmt.Println("Active=", p.Actives())
		fmt.Println("Idles=", p.Idles())
		c := p.Pop()
		connSlice = append(connSlice, c)
	}

	fmt.Println("Conn slice len:", len(connSlice))
	for _, c := range connSlice {
		fmt.Println("Active=", p.Actives())
		fmt.Println("Idles=", p.Idles())
		p.Push(c)
	}
}

func TestPoolAuth(t *testing.T) {
	p := NewPool("10.16.15.121:9991", "1234567890")
	fmt.Println(p.Actives())
	fmt.Println(p.Idles())
	c := p.Pop()
	if c == nil {
		fmt.Println("Pop nil")
		return
	}
	defer p.Push(c)
	// fmt.Println(c.Info())
	fmt.Println(c.Call("SET", "zyh0924", "abcdefghijklmnopqrstuvwxyz"))
	fmt.Println(c.Call("GET", "zyh0924"))
}

func TestWrite(t *testing.T) {
	c, e := Dial("10.16.15.121:9731", "", ConnectTimeout, ReadTimeout, WriteTimeout, false)
	if e != nil {
		println(e.Error())
		return
	}
	defer c.conn.Close()

	// test commands
	key := "zyh0924"
	args := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"}
	fmt.Println(c.SADD(key, args))
	fmt.Println(c.SMEMBERS(key))
	fmt.Println(c.DEL([]string{key}))

	// test pipeline
	c.PipeSend("SET", "a", "zyh")
	c.PipeSend("SET", "b", "zyh")
	c.PipeSend("SET", "c", "zyh")
	fmt.Println(c.PipeExec())

	// test transaction
	c.MULTI()
	c.TransSend("SET", "a", "zyh2")
	c.TransSend("SET", "b", "zyh3")
	fmt.Println(c.TransExec())
}
