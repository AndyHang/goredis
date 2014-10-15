package msgRedis

import (
	"fmt"
	"strconv"
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
			pools[addrPass[0]] = NewPool(addrPass[0], addrPass[1])
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
	if _, ok := mp.pools[addr]; !ok {
		fmt.Println("[PopByAddr] invalid address:" + addr)
		return nil
	}
	return mp.pools[addr].Pop()
}

func (mp *MultiPool) PushByAddr(addr string, c *Conn) {
	if _, ok := mp.pools[addr]; !ok {
		fmt.Println("[PushByAddr] invalid address:" + addr)
		return
	}
	mp.pools[addr].Push(c)
}

// sum(key)/len(pools)
func (mp *MultiPool) PopByKey(key string) *Conn {
	addr := mp.servers[Sum(key)/len(mp.pools)]
	if _, ok := mp.pools[addr]; !ok {
		fmt.Println("[PopByKey] invalid address:" + addr)
		return nil
	}
	return mp.pools[addr].Pop()
}

func (mp *MultiPool) PushByKey(key string, c *Conn) {
	addr := mp.servers[Sum(key)/len(mp.pools)]
	if _, ok := mp.pools[addr]; !ok {
		fmt.Println("[PushByKey] invalid address:" + addr)
		return
	}
	mp.pools[addr].Push(c)
}

const (
	// MaxIdleNum     = 20
	// MaxActiveNum   = 20
	MaxConnNum     = 50
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

	CallNum int64
	callMu  sync.RWMutex

	CallConsume map[string]int
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
			// fmt.Println("[Pop] in case")
			if time.Now().Unix()-c.lastActiveTime > MaxIdleSeconds {
				if c.IsAlive() {
					p.mu.Lock()
					p.IdleNum--
					p.ActiveNum++
					p.mu.Unlock()
					break PopLoop
				}
				c.Close()
				p.mu.Lock()
				p.IdleNum--
				p.mu.Unlock()
				fmt.Println("[Pop] lastActiveTime exceed 30s")
				break
			}
			// if !c.IsAlive() {
			// 	p.mu.Lock()
			// 	p.IdleNum--
			// 	p.mu.Unlock()
			// 	c.Close()
			// 	fmt.Println("[Pop] not alive")
			// 	break
			// }
			p.mu.Lock()
			p.IdleNum--
			p.ActiveNum++
			p.mu.Unlock()
			break PopLoop
		default:
			// fmt.Println("[Pop] in default")
			p.mu.RLock()
			if p.IdleNum+p.ActiveNum >= MaxConnNum {
				p.mu.RUnlock()
				// fmt.Println("waiting................")
				if waitSeconds <= 0 {
					break PopLoop
				}
				waitSeconds--
				fmt.Println("[Pop] max wait 1s")
				time.Sleep(1e9)
				break
			}
			p.mu.RUnlock()

			p.mu.Lock()
			p.ActiveNum++
			p.mu.Unlock()
			c, e := Dial(p.Address, p.Password, ConnectTimeout, ReadTimeout, WriteTimeout, true, p)
			if e != nil {
				p.mu.Lock()
				p.ActiveNum--
				p.mu.Unlock()
				fmt.Println(e.Error())
				break PopLoop
			}

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
	// if !c.IsAlive() {
	// 	p.mu.Lock()
	// 	p.ActiveNum--
	// 	p.mu.Unlock()
	// 	c.Close()
	// 	fmt.Println("[Push] not alive")
	// 	return
	// }
	select {
	case p.ClientPool <- c:
		p.mu.Lock()
		p.IdleNum++
		p.ActiveNum--
		p.mu.Unlock()
		// fmt.Println("[Push] success")
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

// 返回string，根据需要可能会修改返回值类型，如果info包含其他信息
func (p *Pool) PoolInfo() string {
	var IdleN, ActiveN int
	p.mu.RLock()
	IdleN = p.IdleNum
	ActiveN = p.IdleNum
	p.mu.RUnlock()
	return "ActiveNum=" + strconv.Itoa(ActiveN) + "\n IdleNum=" + strconv.Itoa(IdleN) + " \n"
}

func (p *Pool) QPS() int64 {
	var n int64 = 0
	p.callMu.RLock()
	n = p.CallNum
	p.callMu.RUnlock()

	time.Sleep(time.Second)

	p.callMu.RLock()
	n = p.CallNum - n
	p.callMu.RUnlock()
	return n
}

func (p *Pool) QPSAvg() int64 {
	var n int64 = 0
	qps := make([]int64, 4)
	p.callMu.RLock()
	n = p.CallNum
	p.callMu.RUnlock()

	for i := 0; i < 3; i++ {
		time.Sleep(time.Second)
		p.callMu.RLock()
		qps[i] = p.CallNum - n
		n = p.CallNum
		p.callMu.RUnlock()
	}
	qps[3] = (qps[0] + qps[1] + qps[2]) / 3
	return qps[3]
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
