package msgRedis

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DefaultMaxConnNumber  = 50
	DefaultMaxIdleSeconds = 28
)

// include multi redis server's connection pool
type MultiPool struct {
	pools   map[string]*Pool
	servers []string
}

//
func NewMultiPool(addresses []string, maxConnNum int, maxIdleSeconds int64) *MultiPool {
	pools := make(map[string]*Pool, len(addresses))
	for _, addr := range addresses {
		addrPass := strings.Split(addr, "@")
		if len(addrPass) == 2 {
			// redis need auth
			pools[addr] = NewPool(addrPass[0], addrPass[1], maxConnNum, maxIdleSeconds)
		} else if len(addrPass) == 1 {
			// redis do not need auth
			pools[addr] = NewPool(addrPass[0], "", maxConnNum, maxIdleSeconds)
		} else {
			fmt.Println("invalid address format:should 1.1.1.1:1100 or 1.1.1.1:1100@123")
		}
	}

	return &MultiPool{
		pools:   pools,
		servers: addresses,
	}
}

func (mp *MultiPool) AddPool(address string, maxConnNum int, maxIdleSeconds int64) bool {
	addrPass := strings.Split(address, "@")
	if len(addrPass) == 2 {
		// redis need auth
		mp.pools[address] = NewPool(addrPass[0], addrPass[1], maxConnNum, maxIdleSeconds)
	} else if len(addrPass) == 1 {
		// redis do not need auth
		mp.pools[address] = NewPool(addrPass[0], "", maxConnNum, maxIdleSeconds)
	} else {
		fmt.Println("invalid address format:should 1.1.1.1:1100 or 1.1.1.1:1100@123")
		return false
	}
	mp.servers = append(mp.servers, address)
	return true
}

// 暂不支持
func (mp *MultiPool) DelPool() bool {
	return false
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

func (mp *MultiPool) Push(c *Conn) {
	if c == nil {
		return
	}
	addr := c.Address
	if _, ok := mp.pools[addr]; !ok {
		fmt.Println("[Push] invalid address:" + addr)
		c.Close()
		return
	}
	mp.pools[addr].Push(c)
}

func (mp *MultiPool) Info() {
	for _, p := range mp.pools {
		fmt.Println(p.PoolInfo())
	}
}

const (
	// MaxIdleNum     = 20
	// MaxActiveNum   = 20
	GMaxConnNum     = 50
	GMaxIdleSeconds = 28
)

// connection pool of only one redis server
type Pool struct {
	Address        string
	Password       string
	IdleNum        int
	ActiveNum      int
	MaxConnNum     int
	MaxIdleSeconds int64

	ClientPool chan *Conn
	mu         sync.RWMutex

	CallNum int64
	callMu  sync.RWMutex

	ScriptMap map[string]string

	CallConsume map[string]int // 命令消耗时长
}

func NewPool(address, password string, maxConnNum int, maxIdleSeconds int64) *Pool {
	return &Pool{
		Address:        address,
		Password:       password,
		IdleNum:        0,
		ActiveNum:      0,
		MaxConnNum:     maxConnNum,
		MaxIdleSeconds: maxIdleSeconds,
		ClientPool:     make(chan *Conn, maxConnNum),
		ScriptMap:      make(map[string]string, 1),
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
			if time.Now().Unix()-c.lastActiveTime > p.MaxIdleSeconds {
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
				fmt.Println("[Pop] lastActiveTime exceed maxIdleSeconds")
				break
			}
			p.mu.Lock()
			p.IdleNum--
			p.ActiveNum++
			p.mu.Unlock()
			break PopLoop
		default:
			p.mu.RLock()
			if p.IdleNum+p.ActiveNum >= p.MaxConnNum {
				p.mu.RUnlock()
				fmt.Println("waiting................")
				if waitSeconds <= 0 {
					break PopLoop
				}
				waitSeconds--
				fmt.Println("[Pop] max wait 1s")
				time.Sleep(1e9)
				break
			}
			p.mu.RUnlock()

			//
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
		// p.mu.Lock()
		// p.ActiveNum--
		// p.mu.Unlock()
		fmt.Println("[Push] c == nil")
		return
	}

	if c.err != nil {
		c.Close()
		p.mu.Lock()
		p.ActiveNum--
		p.mu.Unlock()
		return
	}

	select {
	case p.ClientPool <- c:
		// fmt.Println("Push in Channel")
		p.mu.Lock()
		p.IdleNum++
		p.ActiveNum--
		p.mu.Unlock()
		// fmt.Println("[Push] success")
	default:
		c.Close()
		p.mu.Lock()
		p.ActiveNum--
		p.mu.Unlock()
		// fmt.Println("[Push] discard")
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
	ActiveN = p.ActiveNum
	p.mu.RUnlock()
	return "Address=" + p.Address + "    ActiveNum=" + strconv.Itoa(ActiveN) +
		"    IdleNum=" + strconv.Itoa(IdleN) + "    ChannelLen=" + strconv.Itoa(len(p.ClientPool))
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

func (p *Pool) AddScriptSha1(name, script string) {
	p.mu.Lock()
	p.ScriptMap[name] = script
	p.mu.Unlock()
}

func (p *Pool) DelScriptSha1(name string) {
	p.mu.Lock()
	delete(p.ScriptMap, name)
	p.mu.Unlock()
}

func (p *Pool) GetScriptSha1(name string) string {
	sha1 := ""
	p.mu.RLock()
	if _, ok := p.ScriptMap[name]; ok {
		sha1 = p.ScriptMap[name]
	}
	p.mu.RUnlock()
	return sha1
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
