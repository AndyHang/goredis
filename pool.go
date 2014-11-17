// Copyright 2014 zhaiyuhang. All rights reserved.
package goredis

import (
	"encoding/json"
	"strings"
	"sync"
	"time"
)

var debug = true

const (
	DefaultMaxConnNumber  = 50
	DefaultMaxIdleNumber  = 25
	DefaultMaxIdleSeconds = 28
)

// include multi redis server's connection pool
type MultiPool struct {
	pools   map[string]*Pool
	servers []string
	mu      sync.RWMutex
}

//
func NewMultiPool(addresses []string, maxConnNum, maxIdleNum int, maxIdleSeconds int64) *MultiPool {
	pools := make(map[string]*Pool, len(addresses))
	for _, addr := range addresses {
		addrPass := strings.Split(addr, ":")
		if len(addrPass) == 3 {
			// redis need auth
			pools[addr] = NewPool(addrPass[0]+":"+addrPass[1], addrPass[2], maxConnNum, maxIdleNum, maxIdleSeconds)
		} else if len(addrPass) == 2 {
			// redis do not need auth
			pools[addr] = NewPool(addr, "", maxConnNum, maxIdleNum, maxIdleSeconds)
		} else {
			Debug("invalid address format:should 1.1.1.1:1100 or 1.1.1.1:1100:123", addr)
		}
	}

	return &MultiPool{
		pools:   pools,
		servers: addresses,
	}
}

func (mp *MultiPool) AddPool(address string, maxConnNum, maxIdleNum int, maxIdleSeconds int64) bool {
	addrPass := strings.Split(address, ":")
	if len(addrPass) == 3 {
		// redis need auth
		mp.mu.Lock()
		mp.pools[address] = NewPool(addrPass[0]+":"+addrPass[1], addrPass[2], maxConnNum, maxIdleNum, maxIdleSeconds)
		mp.mu.Unlock()
	} else if len(addrPass) == 2 {
		// redis do not need auth
		mp.mu.Lock()
		mp.pools[address] = NewPool(address, "", maxConnNum, maxIdleNum, maxIdleSeconds)
		mp.mu.Unlock()
	} else {
		Debug("invalid address format:should 1.1.1.1:1100 or 1.1.1.1:1100:123", address)
		return false
	}
	mp.mu.Lock()
	mp.servers = append(mp.servers, address)
	mp.mu.Unlock()
	return true
}

func (mp *MultiPool) DelPool(address string) {
	mp.mu.Lock()
	delete(mp.pools, address)
	delIndex := -1
	for index, addr := range mp.servers {
		if addr == address {
			delIndex = index
			break
		}
	}
	if delIndex != -1 {
		mp.servers = append(mp.servers[:delIndex], mp.servers[delIndex+1:]...)
	}
	mp.mu.Unlock()
}

func (mp *MultiPool) ReplacePool(src, dst string, maxConnNum, maxIdleNum int, maxIdleSeconds int64) bool {
	mp.mu.RLock()
	_, ok := mp.pools[src]
	delete(mp.pools, src)
	mp.mu.RUnlock()
	if !ok {
		Debug("src not exists in the pool", src)
		return false
	}

	addrPass := strings.Split(dst, ":")
	if len(addrPass) == 3 {
		// redis need auth
		mp.mu.Lock()
		mp.pools[dst] = NewPool(addrPass[0]+":"+addrPass[1], addrPass[2], maxConnNum, maxIdleNum, maxIdleSeconds)
		mp.mu.Unlock()
	} else if len(addrPass) == 2 {
		// redis do not need auth
		mp.mu.Lock()
		mp.pools[dst] = NewPool(dst, "", maxConnNum, maxIdleNum, maxIdleSeconds)
		mp.mu.Unlock()
	} else {
		Debug("invalid address format:should 1.1.1.1:1100 or 1.1.1.1:1100:123", dst)
		return false
	}
	mp.mu.Lock()
	for _, server := range mp.servers {
		if server == src {
			server = dst
		}
	}
	mp.mu.Unlock()
	return true
}

// get conn by address directly
func (mp *MultiPool) PopByAddr(addr string) *Conn {
	mp.mu.RLock()
	pool, ok := mp.pools[addr]
	mp.mu.RUnlock()
	if !ok {
		Debug("[PopByAddr] invalid", addr)
		return nil
	}
	return pool.Pop()
}

func (mp *MultiPool) PushByAddr(addr string, c *Conn) {
	mp.mu.RLock()
	pool, ok := mp.pools[addr]
	mp.mu.RUnlock()
	if !ok {
		Debug("[PushByAddr] invalid", addr)
		return
	}
	pool.Push(c)
}

// sum(key)/len(pools)
func (mp *MultiPool) PopByKey(key string) *Conn {
	mp.mu.RLock()
	addr := mp.servers[Sum(key)%len(mp.pools)]
	pool, ok := mp.pools[addr]
	mp.mu.RUnlock()

	if !ok {
		Debug("[PopByKey] invalid", addr)
		return nil
	}
	return pool.Pop()
}

func (mp *MultiPool) PushByKey(key string, c *Conn) {
	mp.mu.RLock()
	addr := mp.servers[Sum(key)%len(mp.pools)]
	pool, ok := mp.pools[addr]
	mp.mu.RUnlock()
	if !ok {
		Debug("[PushByKey] invalid", addr)
		return
	}
	pool.Push(c)
}

func (mp *MultiPool) Push(c *Conn) {
	if c == nil {
		return
	}
	addr := c.Address
	mp.mu.RLock()
	pool, ok := mp.pools[addr]
	mp.mu.RUnlock()
	if !ok {
		Debug("[Push] invalid", addr)
		return
	}
	pool.Push(c)
}

func (mp *MultiPool) Info() string {
	var jsonLock sync.Mutex
	var wait sync.WaitGroup
	mp.mu.RLock()
	jsonSlice := make([]*PoolInfo, 0, len(mp.pools))
	for _, p := range mp.pools {
		wait.Add(1)
		go func(p *Pool) {
			info := p.Info()
			jsonLock.Lock()
			jsonSlice = append(jsonSlice, info)
			jsonLock.Unlock()
			wait.Done()
		}(p)
	}
	mp.mu.RUnlock()
	wait.Wait()

	responseJson, _ := json.Marshal(jsonSlice)
	return string(responseJson)
}

// connection pool of only one redis server
type Pool struct {
	Address  string
	Password string

	// 统计信息
	IdleNum         int
	ActiveNum       int
	MaxConnNum      int
	MaxIdleNum      int
	CreateNum       int
	CreateFailedNum int
	WaitTimeoutNum  int
	PingErrNum      int
	CallNetErrNum   int

	MaxIdleSeconds int64

	ClientPool chan *Conn
	mu         sync.RWMutex

	CallNum int64
	callMu  sync.RWMutex

	ScriptMap map[string]string

	CallConsume map[string]int // 命令消耗时长
}

func NewPool(address, password string, maxConnNum, maxIdleNum int, maxIdleSeconds int64) *Pool {
	return &Pool{
		Address:        address,
		Password:       password,
		IdleNum:        0,
		ActiveNum:      0,
		CreateNum:      0,
		MaxConnNum:     maxConnNum,
		MaxIdleNum:     maxIdleNum,
		MaxIdleSeconds: maxIdleSeconds,
		ClientPool:     make(chan *Conn, maxConnNum),
		ScriptMap:      make(map[string]string, 1),
	}
}

// TODO: add timeout
func (p *Pool) Pop() *Conn {
	var waitSeconds = 50
	var c *Conn
PopLoop:
	for {
		select {
		case c = <-p.ClientPool:
			if time.Now().Unix()-c.lastActiveTime > p.MaxIdleSeconds {
				if c.IsAlive() {
					// 标记当前连接为正在使用
					c.Lock()
					c.isIdle = false
					c.Unlock()
					// 更新连接计数
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
				break
			}
			// 标记当前连接为正在使用
			c.Lock()
			c.isIdle = false
			c.Unlock()

			p.mu.Lock()
			p.IdleNum--
			p.ActiveNum++
			p.mu.Unlock()
			break PopLoop
		default:
			p.mu.RLock()
			if p.IdleNum+p.ActiveNum >= p.MaxConnNum {
				p.mu.RUnlock()
				if waitSeconds <= 0 {
					p.mu.Lock()
					p.WaitTimeoutNum++
					p.mu.Unlock()
					Debug("waiting exceed time get conn failed ", p.Address)
					break PopLoop
				}
				waitSeconds--
				time.Sleep(2e8)
				break
			}
			p.mu.RUnlock()

			//
			p.mu.Lock()
			p.ActiveNum++
			p.CreateNum++
			p.mu.Unlock()
			c, e := Dial(p.Address, p.Password, ConnectTimeout, ReadTimeout, WriteTimeout, true, p)
			if e != nil {
				p.mu.Lock()
				p.ActiveNum--
				p.CreateFailedNum++
				p.mu.Unlock()
				Debug(e.Error(), p.Address)
				break PopLoop
			}
			// 标记当前连接为正在使用
			c.Lock()
			c.isIdle = false
			c.Unlock()

			p.Push(c)
		}
	}
	return c
}

func (p *Pool) Push(c *Conn) {
	if c == nil {
		Debug("[Push] c == nil", p.Address)
		return
	}

	// 如果已经在连接池里，就不能再push进pool中
	c.Lock()
	if c.isIdle {
		c.Unlock()
		return
	}
	c.isIdle = true
	c.Unlock()

	// 如果连接网络出错，直接丢掉
	if c.err != nil {
		p.mu.Lock()
		p.ActiveNum--
		p.mu.Unlock()
		c.Close()
		return
	}

	p.mu.RLock()
	idleNum := p.IdleNum
	p.mu.RUnlock()
	if idleNum > p.MaxIdleNum {
		p.mu.Lock()
		p.ActiveNum--
		p.mu.Unlock()
		c.Close()
		return
	}

	select {
	case p.ClientPool <- c:
		p.mu.Lock()
		p.IdleNum++
		p.ActiveNum--
		p.mu.Unlock()
	default:
		p.mu.Lock()
		p.ActiveNum--
		p.mu.Unlock()
		c.Close()
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

type PoolInfo struct {
	Address         string
	IdleNum         int
	ActiveNum       int
	CreateNum       int
	TimeoutNum      int
	CreateFailedNum int
	CallNetErrNum   int
	PingErrNum      int
	Qps             int64
}

// 返回string，根据需要可能会修改返回值类型，如果info包含其他信息
func (p *Pool) Info() *PoolInfo {
	p.mu.RLock()
	IdleN := p.IdleNum
	ActiveN := p.ActiveNum
	CreateN := p.CreateNum
	TimeoutN := p.WaitTimeoutNum
	CreateFailedN := p.CreateFailedNum
	CallNetErrN := p.CallNetErrNum
	PingErrN := p.PingErrNum
	p.mu.RUnlock()

	qps := p.QPS()
	poolInfo := &PoolInfo{
		Address:         p.Address,
		IdleNum:         IdleN,
		ActiveNum:       ActiveN,
		CreateNum:       CreateN,
		TimeoutNum:      TimeoutN,
		CreateFailedNum: CreateFailedN,
		CallNetErrNum:   CallNetErrN,
		PingErrNum:      PingErrN,
		Qps:             qps,
	}

	return poolInfo
	// v, _ := json.Marshal(poolInfo)
	// return v
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

// 当前时间
func Now() string {
	return time.Now().Format("2006-01-02 15:04:05 ")
}

func Debug(info, address string) {
	if debug {
		println(Now() + info + "|addr=" + address)
	}
}
