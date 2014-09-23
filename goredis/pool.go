package msgredis

import (
	"sync"
	"time"
)

const (
	MaxIdleNum     = 10
	MaxActiveNum   = 10
	MaxConnNum     = 20
	MaxIdleSeconds = 28
)

type Pool struct {
	IdleNum    int
	ActiveNum  int
	ClientPool chan *Conn
	mu         sync.RWMutex
}

func NewPool(addresses) *Pool {
	return &Pool{
		IdleNum:    0,
		ActiveNum:  0,
		ClientPool: make(chan *Conn, MaxIdleNum),
	}
}

func (p *Pool) Dial() *Conn {
	p.mu.RLock()
	if p.IdleNum+p.ActiveNum > MaxConnNum {
		p.mu.RUnlock()
		return nil
	}

}

// TODO: ping一次，看是否有效
func (p *Pool) Pop() *Conn {
	var c *Conn
	select {
	case c = <-p.ClientPool:
		if time.Now().Unix()-c.lastActiveTime > MaxIdleSeconds {
			c.Close()
			p.IdleNum--
			break
		}
		p.mu.Lock()
		p.IdleNum--
		p.ActiveNum++
		p.mu.Unlock()
	default:
	}
	return c
}

func (p *Pool) Push(c *Conn) {

}
