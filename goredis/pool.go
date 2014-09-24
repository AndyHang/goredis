package msgredis

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

const ()

// include multi redis server's connection pool
type MultiPool struct {
	pools   map[string]*Pool
	servers []string
}

//
func NewMultiPool(addresses []string) *MultiPool {
	pools := make(map[string]*Pool, len(addresses))
	for _, addr := range addresses {
		addrPass := strings.Split(addr, "@")
		if len(addrPass) == 2 {
			// redis need auth
			pools[addr] = NewPool(addrPass[0], addrPass[1])
		} else if len(addrPass) == 1 {
			// redis do not need auth
			pools[addr] = NewPool(addrPass[0], "")
		} else {
			fmt.Println("invalid address format:should 1.1.1.1:1100 or 1.1.1.1:1100@123")
		}
	}

	return &MultiPool{
		pools:   pools,
		servers: addresses,
	}
}

// get conn by address directly
func (mp *MultiPool) PopByAddr(addr string) *Conn {
	if _, ok := mp.pools[addr]; ok {
		return nil
	}
	return mp.pools[addr].Pop()
}

func (mp *MultiPool) PushByAddr(addr string, c *Conn) {
	if _, ok := mp.pools[addr]; ok {
		return
	}
	mp.pools[addr].Push(c)
}

// sum(key)/len(pools)
func (mp *MultiPool) PopByKey(key string) *Conn {
	addr := mp.servers[Sum(key)/len(mp.pools)]
	if _, ok := mp.pools[addr]; ok {
		return nil
	}
	return mp.pools[addr].Pop()
}

func (mp *MultiPool) PushByKey(key string, c *Conn) {
	addr := mp.servers[Sum(key)/len(mp.pools)]
	if _, ok := mp.pools[addr]; ok {
		return
	}
	mp.pools[addr].Push(c)
}

const (
	// MaxIdleNum     = 20
	// MaxActiveNum   = 20
	MaxConnNum     = 20
	MaxIdleSeconds = 28
)

// connection pool of only one redis server
type Pool struct {
	Address    string
	Password   string
	IdleNum    int
	ActiveNum  int
	ClientPool chan *Conn
	mu         sync.RWMutex
}

func NewPool(address, password string) *Pool {
	return &Pool{
		Address:    address,
		Password:   password,
		IdleNum:    0,
		ActiveNum:  0,
		ClientPool: make(chan *Conn, MaxConnNum),
	}
}

// TODO: add timeout
func (p *Pool) Pop() *Conn {
	var waitSeconds = 5
	var c *Conn
PopLoop:
	for {
		select {
		case c = <-p.ClientPool:
			fmt.Println("[Pop] in case")
			if time.Now().Unix()-c.lastActiveTime > MaxIdleSeconds {
				c.Close()
				p.mu.Lock()
				p.IdleNum--
				p.mu.Unlock()
				fmt.Println("[Pop] lastActiveTime exceed 30s")
				break
			}
			if !c.IsAlive() {
				p.mu.Lock()
				p.IdleNum--
				p.mu.Unlock()
				c.Close()
				fmt.Println("[Pop] not alive")
				break
			}
			p.mu.Lock()
			p.IdleNum--
			p.ActiveNum++
			p.mu.Unlock()
			break PopLoop
		default:
			fmt.Println("[Pop] in default")
			p.mu.RLock()
			if p.IdleNum+p.ActiveNum >= MaxConnNum {
				p.mu.RUnlock()
				if waitSeconds <= 0 {
					break PopLoop
				}
				waitSeconds--
				fmt.Println("[Pop] max wait 1s")
				time.Sleep(1e9)
				break
			}
			p.mu.RUnlock()
			c, e := Dial(p.Address, p.Password, ConnectTimeout, ReadTimeout, WriteTimeout, true)
			if e != nil {
				fmt.Println("dial conn error")
				break PopLoop
			}
			p.mu.Lock()
			p.ActiveNum++
			p.mu.Unlock()
			p.Push(c)
		}
	}
	return c
}

func (p *Pool) Push(c *Conn) {
	if c == nil {
		fmt.Println("[Push] c == nil")
		return
	}
	if !c.IsAlive() {
		p.mu.Lock()
		p.ActiveNum--
		p.mu.Unlock()
		c.Close()
		fmt.Println("[Push] not alive")
		return
	}
	select {
	case p.ClientPool <- c:
		p.mu.Lock()
		p.IdleNum++
		p.ActiveNum--
		p.mu.Unlock()
		fmt.Println("[Push] success")
	default:
		c.Close()
		fmt.Println("[Push] discard")
		// discard
	}
}

func (p *Pool) Actives() int {
	var n int
	p.mu.RLock()
	n = p.ActiveNum
	p.mu.RUnlock()
	return n
}

func (p *Pool) Idles() int {
	var n int
	p.mu.RLock()
	n = p.IdleNum
	p.mu.RUnlock()
	return n
}

// 哈希算法
func Sum(key string) int {
	var hash uint32 = 0
	for i := 0; i < len(key); i++ {
		hash += uint32(key[i])
		hash += (hash << 10)
		hash ^= (hash >> 6)
	}
	hash += (hash << 3)
	hash ^= (hash >> 11)
	hash += (hash << 15)
	return int(hash)
}
