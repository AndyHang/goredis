package msgRedis

import (
	// "fmt"
	"runtime"
	// "strconv"
	"sync"
	"testing"
	"time"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func TestMultiPool(t *testing.T) {
	addresses := []string{"10.16.15.121:9731", "10.16.15.121:9991@1234567890"}
	addr := "10.16.15.121:9991"
	mp := NewMultiPool(addresses)
	var g sync.WaitGroup
	for i := 0; i < 10; i++ {
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
			c.CallN(3, "PING")
			mp.PushByAddr(addr, c)
		}()
	}
	g.Wait()
}
