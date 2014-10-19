package msgRedis

import (
	"fmt"
	"runtime"
	// "strconv"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func TestMultiPop(t *testing.T) {
	addresses := []string{"10.16.15.121:9731", "10.16.15.121:9991@1234567890"}
	// addr := "10.16.15.121:9991@1234567890"
	addr := "10.16.15.121:9731"
	mp := NewMultiPool(addresses, 3, 9)
	// fmt.Println(mp.AddPool("10.16.15.121:9901", 10, 60))
	go func() {
		for i := 0; i < 29; i++ {
			mp.Info()
			time.Sleep(time.Second)
		}
	}()

	// }
	for i := 0; i < 2; i++ {
		c := mp.PopByAddr(addr)
		if c == nil {
			t.Error("c==nil....................")
			return
		}

		fmt.Println(c.SET("key", "value"))
		mp.Push(c)
		time.Sleep(time.Duration(11 * time.Second))
	}

}

func TestMultiPool(t *testing.T) {
	addresses := []string{"10.16.15.121:9731", "10.16.15.121:9991@1234567890"}
	addr := "10.16.15.121:9991@1234567890"
	mp := NewMultiPool(addresses, 20, 20)
	fmt.Println(mp.AddPool("10.16.15.121:9901", 10, 60))
	go func() {
		for i := 0; i < 29; i++ {
			mp.Info()
			time.Sleep(time.Second)
		}
	}()

	// for i := 0; i < 20; i++ {
	// 	// time.Sleep(1000)
	// 	c := mp.PopByAddr(addr)
	// 	if c == nil {
	// 		t.Error("c==nil....................")
	// 		return
	// 	}
	// 	// c.PipeSend("set", strconv.Itoa(i), strconv.Itoa(i))
	// 	// c.PipeExec()
	// 	// fmt.Println("PING")
	// 	c.CallN(3, "PING")
	// 	time.Sleep(time.Second * 2)
	// 	mp.PushByAddr(addr, c)

	// }

	var g sync.WaitGroup
	for i := 0; i < 50; i++ {
		g.Add(1)
		go func() {
			defer g.Done()
			time.Sleep(1000)
			c := mp.PopByAddr(addr)
			if c == nil {
				t.Error("c==nil....................")
				return
			}
			// c.PipeSend("set", strconv.Itoa(i), strconv.Itoa(i))
			// c.PipeExec()
			// fmt.Println("PING")
			n := rand.Intn(10)
			c.CallN(3, "PING")
			time.Sleep(time.Duration(n * 1e9))
			// mp.PushByAddr(addr, c)
			mp.Push(c)
		}()
	}
	g.Wait()
	time.Sleep(30e9)
}
