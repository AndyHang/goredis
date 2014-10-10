package msgredis

import (
	"fmt"
	"testing"
	"time"
)

func TestCallN(t *testing.T) {
	p := NewPool("10.16.15.121:9731", "")
	c := p.Pop()
	fmt.Println(c.CallN(2, "HMGET", "a", "a"))
}

func TestPipeSmallBuffer(t *testing.T) {
	c, e := Dial("10.16.15.121:9731", "", ConnectTimeout, ReadTimeout, WriteTimeout, false, nil)
	if e != nil {
		println(e.Error())
		return
	}
	defer c.conn.Close()
	// test pipeline
	for i := 0; i < 1000; i++ {
		c.PipeSend("INCR", "Zincr")
	}
	fmt.Println(c.PipeExec())
}

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
	c, e := Dial("10.16.15.121:9731", "", ConnectTimeout, ReadTimeout, WriteTimeout, false, nil)
	if e != nil {
		println(e.Error())
		return
	}
	defer c.conn.Close()

	// test commands
	key := "zyh1008"
	args := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"}
	fmt.Println(c.SET(key, "zzzaaa"))
	fmt.Println(c.GET(key))
	return

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
func TestCommands(t *testing.T) {
	c, e := Dial("10.16.15.121:9731", "", ConnectTimeout, ReadTimeout, WriteTimeout, false, nil)
	if e != nil {
		println(e.Error())
		return
	}
	defer c.conn.Close()

	fmt.Println(c.KEYS("*dadfjiii*"))

	key := "zyh1009"
	fmt.Println("STRINGS.************************STRINGS**********************.STRINGS")
	fmt.Println(c.Call("FLUSHDB"))
	fmt.Println(c.GET(key))
	fmt.Println(c.SET(key, "zyh1009"))
	fmt.Println(c.GET(key))
	fmt.Println(c.SETBIT(key, 10, 0))
	fmt.Println(c.GETBIT(key, 10))
	fmt.Println(c.GETRANGE(key, 0, 3))
	fmt.Println(c.MGET([]string{key, "key5"}))
	kv := make(map[string]string)
	kv["key2"] = "key2a"
	kv["key3"] = "key3a"
	fmt.Println(c.MSET(kv))
	fmt.Println(c.MSETNX(kv))
	fmt.Println(c.PSETEX(key, 100000, "abc"))
	fmt.Println(c.SETEX(key, 110, "abc1"))
	fmt.Println(c.SETRANGE(key, 1, "aaaa"))
	fmt.Println(c.STRLEN(key))

	// hashes
	key = "hashesZYH"
	fields := []string{"f1", "f2"}
	fmt.Println("HASHES.************************HASHES**********************.HASHES")
	fmt.Println(c.HDEL(key, fields))
	fmt.Println(c.HEXISTS(key, fields[0]))
	fmt.Println(c.HEXISTS(key, "noexists"))
	fmt.Println(c.HGET(key, fields[0]))
	fmt.Println(c.HGET(key, "noexists"))
	fmt.Println(c.HGETALL(key))
	fmt.Println(c.HGETALL("noexists"))
	fmt.Println(c.HINCRBY(key, "field", 1))
	fmt.Println(c.HINCRBYFLOAT(key, "field1", 1.1))
	fmt.Println(c.HKEYS(key))
	fmt.Println(c.HLEN(key))
	fmt.Println(c.HMGET(key, []string{key, "noexists"}))

	kv1 := make(map[string]interface{})
	kv1["key2"] = "keys2"
	kv1["key3"] = "keyss3"

	fmt.Println(c.HMSET(key, kv1))
	fmt.Println(c.HSET(key, "fieldFloat", 1.5))
	fmt.Println(c.HSETNX(key, "fieldFloat11", 1.6))
	fmt.Println(c.HVALS(key))

	fmt.Println("LISTS.***********************LISTS***********************.LISTS")
	key = "listsZYH"
	key1 := "lists1"
	fmt.Println(c.BLPOP([]string{"key11", "key21"}, 2))
	fmt.Println(c.BRPOP([]string{"key11", "key12"}, 2))
	fmt.Println(c.LPUSH(key, []string{"a", "b", "c"}))
	fmt.Println(c.RPUSH(key, []string{"x", "y", "z", "a"}))
	fmt.Println(c.BRPOPLPUSH(key, key1, 1))
	fmt.Println(c.LINDEX(key, 4))
	fmt.Println(c.LINSERT(key, "after", "a", "d"))
	fmt.Println(c.LLEN(key))
	fmt.Println(c.LPUSHX("noexists", "value"))
	fmt.Println(c.LRANGE(key, 0, 10))
	fmt.Println(c.LREM(key, -1, "a"))
	fmt.Println(c.LSET(key, 0, "0"))
	fmt.Println(c.LTRIM(key, 1, -1))
	fmt.Println(c.RPOP(key))
	fmt.Println(c.RPOPLPUSH(key, key1))
	fmt.Println(c.RPUSHX(key, "value"))

	fmt.Println("SETS.**************************SETS*********************.SETS")

	fmt.Println("SORTED SETS.********************SORTED SETS******************.SORTED SETS")
	// key = "listsZYH"
}

func TestQPS(t *testing.T) {
	p := NewPool("10.16.15.121:9731", "")
	for i := 0; i < 50; i++ {
		c := p.Pop()
		go call(c)
	}

	go func() {
		for {
			fmt.Println("QPS:", p.QPS())
		}
	}()

	// select {}
	time.Sleep(11e9)
}

func call(c *Conn) {
	for {
		key := "zyh1009"
		c.Call("FLUSHDB")
		c.GET(key)
		c.SET(key, "zyh1009")
		c.GET(key)
		c.SETBIT(key, 10, 0)
		c.GETBIT(key, 10)
		c.GETRANGE(key, 0, 3)
		c.MGET([]string{key, "key5"})
		kv := make(map[string]string)
		kv["key2"] = "key2a"
		kv["key3"] = "key3a"
		c.MSET(kv)
		c.MSETNX(kv)
		c.PSETEX(key, 100000, "abc")
		c.SETEX(key, 110, "abc1")
		c.SETRANGE(key, 1, "aaaa")
		c.STRLEN(key)

		// hashes
		key = "hashesZYH"
		fields := []string{"f1", "f2"}
		c.HDEL(key, fields)
		c.HEXISTS(key, "noexists")
		c.HGET(key, fields[0])
		c.HGET(key, "noexists")
		c.HGETALL(key)
		c.HGETALL("noexists")
		c.HINCRBY(key, "field", 1)
		c.HINCRBYFLOAT(key, "field1", 1.1)
		c.HKEYS(key)
		c.HLEN(key)
		c.HMGET(key, []string{key, "noexists"})

		kv1 := make(map[string]interface{})
		kv1["key2"] = "keys2"
		kv1["key3"] = "keyss3"

		c.HMSET(key, kv1)
		c.HSET(key, "fieldFloat", 1.5)
		c.HSETNX(key, "fieldFloat11", 1.6)
		c.HVALS(key)

	}

}
