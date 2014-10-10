package msgredis

// all return integer is int64
import (
	"errors"
	"fmt"
	"strings"
)

func (c *Conn) AUTH(password string) (bool, error) {
	v, e := c.Call("AUTH", password)
	if e != nil {
		fmt.Println("AUTH failed:" + e.Error())
		return false, e
	}

	r, ok := v.([]byte)
	if !ok {
		return false, errors.New("invaild response type")
	}

	if len(r) == 2 && r[0] == 'O' && r[1] == 'K' {
		return true, nil
	}
	return false, errors.New("invaild response string:" + string(r))
}

func (c *Conn) IsAlive() bool {
	v, e := c.Call("PING")
	if e != nil {
		return false
	}
	r, ok := v.([]byte)
	if !ok {
		return false
	}
	if len(r) == 4 && r[0] == 'P' && r[1] == 'O' && r[2] == 'N' && r[3] == 'G' {
		return true
	}
	return false
}

func (c *Conn) Info() ([]byte, error) {
	v, e := c.Call("INFO")
	if e != nil {
		return nil, e
	}
	r, ok := v.([]byte)
	if !ok {
		return nil, errors.New("invalid type")
	}
	return r, nil
}

/******************* keys commands *******************/
func (c *Conn) DEL(keys []string) (int64, error) {
	args := make([]interface{}, len(keys))
	for i := 0; i < len(keys); i++ {
		args[i] = keys[i]
	}
	n, e := c.Call("DEL", args...)
	if e != nil {
		return -1, e
	}
	return n.(int64), nil
}

func (c *Conn) DUMP(key string) ([]byte, error) {
	v, e := c.Call("DUMP", key)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrKeyNotExist
	}
	return v.([]byte), nil
}

func (c *Conn) EXISTS(key string) (bool, error) {
	n, e := c.Call("EXISTS", key)
	if e != nil {
		return false, e
	}

	r := n.(int64)
	if r == 1 {
		return true, nil
	}
	return false, nil
}

func (c *Conn) EXPIRE(key string, seconds int) (bool, error) {
	n, e := c.Call("EXPIRE", key, seconds)
	if e != nil {
		return false, e
	}

	r := n.(int64)
	if r == 1 {
		return true, nil
	}
	return false, nil
}

func (c *Conn) EXPIREAT(key string, timestamp int) (bool, error) {
	n, e := c.Call("EXPIREAT", key, timestamp)
	if e != nil {
		return false, e
	}

	r := n.(int64)
	if r == 1 {
		return true, nil
	}
	return false, nil
}

func (c *Conn) KEYS(pattern string) ([][]byte, error) {
	v, e := c.Call("KEYS", pattern)
	if e != nil {
		return nil, e
	}
	members := make([][]byte, len(v.([]interface{})))
	for i, value := range v.([]interface{}) {
		members[i] = value.([]byte)
	}
	return members, nil
}

/******************* sets commands *******************/
func (c *Conn) SADD(key string, values []string) (int64, error) {
	args := make([]interface{}, len(values)+1)
	args[0] = key
	for i := 0; i < len(values); i++ {
		args[i+1] = values[i]
	}
	n, e := c.Call("SADD", args...)
	if e != nil {
		return -1, e
	}
	return n.(int64), nil
}

func (c *Conn) SREM(key string, values []string) (int64, error) {
	args := make([]interface{}, len(values)+1)
	args[0] = key
	for i := 0; i < len(values); i++ {
		args[i+1] = values[i]
	}
	n, e := c.Call("SREM", args...)
	if e != nil {
		return -1, e
	}
	return n.(int64), nil
}

func (c *Conn) SISMEMBER(key, value string) (int64, error) {
	v, e := c.Call("SCARD", key, value)
	if e != nil {
		return 0, e
	}
	return v.(int64), nil
}

func (c *Conn) SMEMBERS(key string) ([][]byte, error) {
	v, e := c.Call("SMEMBERS", key)
	if e != nil {
		return nil, e
	}
	members := make([][]byte, len(v.([]interface{})))
	for i, value := range v.([]interface{}) {
		members[i] = value.([]byte)
	}
	return members, nil
}

// 0说明key不存在
func (c *Conn) SCARD(key string) (int64, error) {
	v, e := c.Call("SCARD", key)
	if e != nil {
		return 0, e
	}
	return v.(int64), nil
}

func (c *Conn) SINTER(keys []string) ([][]byte, error) {
	args := make([]interface{}, len(keys))
	for i := 0; i < len(keys); i++ {
		args[i] = keys[i]
	}
	v, e := c.Call("SINTER", args...)
	if e != nil {
		return nil, e
	}
	members := make([][]byte, len(v.([]interface{})))
	for i, value := range v.([]interface{}) {
		members[i] = value.([]byte)
	}
	return members, nil
}

func (c *Conn) SINTERSTORE(key string, keys []string) (int64, error) {
	args := make([]interface{}, len(keys)+1)
	args[0] = key
	for i := 0; i < len(keys); i++ {
		args[i+1] = keys[i]
	}
	n, e := c.Call("SINTERSTORE", args...)
	if e != nil {
		return -1, e
	}
	return n.(int64), nil
}

func (c *Conn) SDIFF(keys []string) ([][]byte, error) {
	args := make([]interface{}, len(keys))
	for i := 0; i < len(keys); i++ {
		args[i] = keys[i]
	}
	v, e := c.Call("SDIFF", args...)
	if e != nil {
		return nil, e
	}
	members := make([][]byte, len(v.([]interface{})))
	for i, value := range v.([]interface{}) {
		members[i] = value.([]byte)
	}
	return members, nil
}

func (c *Conn) SDIFFSTORE(key string, keys []string) (int64, error) {
	args := make([]interface{}, len(keys)+1)
	args[0] = key
	for i := 0; i < len(keys); i++ {
		args[i+1] = keys[i]
	}
	n, e := c.Call("SDIFFSTORE", args...)
	if e != nil {
		return -1, e
	}
	return n.(int64), nil
}

// TODO:return bool
func (c *Conn) SMOVE(srcKey, desKey, member string) (int64, error) {
	n, e := c.Call("SMOVE", srcKey, desKey, member)
	if e != nil {
		return -1, e
	}
	return n.(int64), nil
}

func (c *Conn) SPOP(key string) ([]byte, error) {
	v, e := c.Call("SPOP", key)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrKeyNotExist
	}
	return v.([]byte), nil
}

func (c *Conn) SRANDMEMBER(key string, count int) ([][]byte, error) {
	if count == 0 {
		v, e := c.Call("SRANDMEMBER", key)
		if e != nil {
			return nil, e
		}
		members := make([][]byte, 1)
		members[0] = v.([]byte)
		return members, nil
	}
	v, e := c.Call("SRANDMEMBER", key)
	if e != nil {
		return nil, e
	}

	if v == nil {
		return nil, ErrKeyNotExist
	}

	members := make([][]byte, len(v.([]interface{})))
	for i, value := range v.([]interface{}) {
		members[i] = value.([]byte)
	}
	return members, nil
}

func (c *Conn) SUNION(keys []string) ([][]byte, error) {
	args := make([]interface{}, len(keys))
	for i := 0; i < len(keys); i++ {
		args[i] = keys[i]
	}
	v, e := c.Call("SUNION", args...)
	if e != nil {
		return nil, e
	}
	members := make([][]byte, len(v.([]interface{})))
	for i, value := range v.([]interface{}) {
		members[i] = value.([]byte)
	}
	return members, nil
}

func (c *Conn) SUNIONSTORE(key string, keys []string) (int64, error) {
	args := make([]interface{}, len(keys)+1)
	args[0] = key
	for i := 0; i < len(keys); i++ {
		args[i+1] = keys[i]
	}
	n, e := c.Call("SUNIONSTORE", args...)
	if e != nil {
		return -1, e
	}
	return n.(int64), nil
}

// available since redis 2.8.0
// func (c *Conn) SSCAN(key string, cursor int) (int64, [][]byte, error) {

// }

/******************* strings commands *******************/
func (c *Conn) APPEND(key, value string) (int64, error) {
	n, e := c.Call("APPEND", key, value)
	if e != nil {
		return -1, e
	}
	return n.(int64), e
}

func (c *Conn) BITCOUNT(key string) (int64, error) {
	n, e := c.Call("BITCOUNT", key)
	if e != nil {
		return -1, e
	}
	return n.(int64), e
}

// 2.6.0
func (c *Conn) BITOP(op, dest string, keys []string) (int64, error) {
	args := make([]interface{}, len(keys)+2)
	args[0] = op
	args[1] = dest
	for i := 0; i < len(keys); i++ {
		args[i+2] = keys[i]
	}
	n, e := c.Call("BITOP", args...)
	if e != nil {
		return -1, e
	}
	return n.(int64), e
}

// 2.8.7 TODO
func (c *Conn) BITPOS() {}

func (c *Conn) DECR(key string) (int64, error) {
	n, e := c.Call("DECR", key)
	if e != nil {
		return -1, e
	}
	return n.(int64), e
}

func (c *Conn) DECRBY(key string, num int) (int64, error) {
	n, e := c.Call("DECRBY", key, num)
	if e != nil {
		return -1, e
	}
	return n.(int64), e
}

func (c *Conn) INCR(key string) (int64, error) {
	n, e := c.Call("INCR", key)
	if e != nil {
		return -1, e
	}
	return n.(int64), e
}

func (c *Conn) INCRBY(key string, num int) (int64, error) {
	n, e := c.Call("INCRBY", key, num)
	if e != nil {
		return -1, e
	}
	return n.(int64), e
}

func (c *Conn) INCRBYFLOAT(key string, f float64) ([]byte, error) {
	n, e := c.Call("INCRBYFLOAT", key, f)
	if e != nil {
		return nil, e
	}
	return n.([]byte), e
}

func (c *Conn) SET(key, value string) ([]byte, error) {
	v, e := c.Call("SET", key, value)
	if e != nil {
		return nil, e
	}
	return v.([]byte), nil
}

// 应该返回interface还是[]byte?
func (c *Conn) GET(key string) ([]byte, error) {
	v, e := c.Call("GET", key)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrKeyNotExist
	}
	return v.([]byte), nil
}

func (c *Conn) GETBIT(key string, pos int) (int64, error) {
	n, e := c.Call("GETBIT", key, pos)
	if e != nil {
		return -1, e
	}
	return n.(int64), e
}

func (c *Conn) GETRANGE(key string, start, end int) ([]byte, error) {
	v, e := c.Call("GETRANGE", key, start, end)
	if e != nil {
		return nil, e
	}
	return v.([]byte), nil
}

func (c *Conn) GETSET(key, value string) ([]byte, error) {
	v, e := c.Call("GETSET", key, value)
	if e != nil {
		return nil, e
	}

	if v == nil {
		return nil, ErrKeyNotExist
	}
	return v.([]byte), nil
}

func (c *Conn) MGET(keys []string) ([]interface{}, error) {
	args := make([]interface{}, len(keys))
	for k, v := range keys {
		args[k] = v
	}
	v, e := c.Call("MGET", args...)
	if e != nil {
		return nil, e
	}
	return v.([]interface{}), nil
}

func (c *Conn) MSET(kv map[string]string) ([]byte, error) {
	args := make([]interface{}, 2*len(kv))
	i := 0
	for k, v := range kv {
		args[i] = k
		args[i+1] = v
		i = i + 2
	}
	v, e := c.Call("MSET", args...)
	if e != nil {
		return nil, e
	}
	return v.([]byte), nil
}

func (c *Conn) MSETNX(kv map[string]string) (int64, error) {
	args := make([]interface{}, 2*len(kv))
	i := 0
	for k, v := range kv {
		args[i] = k
		args[i+1] = v
		i = i + 2
	}
	v, e := c.Call("MSETNX", args...)
	if e != nil {
		return -1, e
	}
	return v.(int64), e
}

func (c *Conn) PSETEX(key string, millonseconds int64, value string) ([]byte, error) {
	v, e := c.Call("PSETEX", key, millonseconds, value)
	if e != nil {
		return nil, e
	}
	return v.([]byte), nil
}

func (c *Conn) SETBIT(key string, pos, value int) (int64, error) {
	n, e := c.Call("SETBIT", key, pos, value)
	if e != nil {
		return -1, e
	}
	return n.(int64), e
}

func (c *Conn) SETEX(key string, seconds int64, value string) ([]byte, error) {
	v, e := c.Call("SETEX", key, seconds, value)
	if e != nil {
		return nil, e
	}
	return v.([]byte), nil
}

func (c *Conn) SETNX(key, value string) (int64, error) {
	v, e := c.Call("SETNX", key, value)
	if e != nil {
		return -1, e
	}
	return v.(int64), e
}

func (c *Conn) SETRANGE(key string, offset int, value string) (int64, error) {
	v, e := c.Call("SETRANGE", key, offset, value)
	if e != nil {
		return -1, e
	}
	return v.(int64), e
}

func (c *Conn) STRLEN(key string) (int64, error) {
	v, e := c.Call("STRLEN", key)
	if e != nil {
		return -1, e
	}
	return v.(int64), e
}

func (c *Conn) SSCAN(key string) {}

/******************* hashes commands *******************/
func (c *Conn) HDEL(key string, fields []string) (int64, error) {
	args := make([]interface{}, len(fields)+1)
	args[0] = key
	for i := 0; i < len(fields); i++ {
		args[i+1] = fields[i]
	}
	n, e := c.Call("HDEL", args...)
	if e != nil {
		return -1, e
	}
	return n.(int64), nil
}

func (c *Conn) HEXISTS(key string, field string) (bool, error) {
	n, e := c.Call("HEXISTS", key, field)
	if e != nil {
		return false, e
	}
	r := n.(int64)
	if r == 1 {
		return true, nil
	}
	return false, nil
}

func (c *Conn) HGET(key string, field string) ([]byte, error) {
	v, e := c.Call("HGET", key, field)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrKeyNotExist
	}
	return v.([]byte), nil
}

func (c *Conn) HGETALL(key string) ([]interface{}, error) {
	v, e := c.Call("HGETALL", key)
	if e != nil {
		return nil, e
	}

	if v == nil {
		return nil, ErrKeyNotExist
	}
	return v.([]interface{}), nil
}

func (c *Conn) HINCRBY(key string, field string, increment int) (int64, error) {
	n, e := c.Call("HINCRBY", key, field, increment)
	if e != nil {
		return -1, e
	}
	return n.(int64), nil
}

func (c *Conn) HINCRBYFLOAT(key string, field string, increment float64) ([]byte, error) {
	n, e := c.Call("HINCRBYFLOAT", key, field, increment)
	if e != nil {
		return nil, e
	}
	return n.([]byte), nil
}

func (c *Conn) HKEYS(key string) ([][]byte, error) {
	v, e := c.Call("HKEYS", key)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrKeyNotExist
	}
	members := make([][]byte, len(v.([]interface{})))
	for i, value := range v.([]interface{}) {
		members[i] = value.([]byte)
	}
	return members, nil
}

func (c *Conn) HLEN(key string) (int64, error) {
	n, e := c.Call("HLEN", key)
	if e != nil {
		return -1, e
	}
	return n.(int64), nil
}

func (c *Conn) HMGET(key string, fields []string) ([]interface{}, error) {
	args := make([]interface{}, len(fields)+1)
	args[0] = key
	for i := 0; i < len(fields); i++ {
		args[i+1] = fields[i]
	}
	v, e := c.Call("HMGET", args...)
	if e != nil {
		return nil, e
	}

	if v == nil {
		return nil, ErrKeyNotExist
	}
	return v.([]interface{}), nil
}

func (c *Conn) HMSET(key string, kv map[string]interface{}) ([]byte, error) {
	args := make([]interface{}, 2*len(kv)+1)
	args[0] = key
	i := 1
	for k, v := range kv {
		args[i] = k
		args[i+1] = v
		i = i + 2
	}
	v, e := c.Call("HMSET", args...)
	if e != nil {
		return nil, e
	}
	return v.([]byte), nil
}

func (c *Conn) HSET(key, field string, value interface{}) (int64, error) {
	n, e := c.Call("HSET", key, field, value)
	if e != nil {
		return -1, e
	}
	return n.(int64), nil
}

func (c *Conn) HSETNX(key, field string, value interface{}) (int64, error) {
	n, e := c.Call("HSETNX", key, field, value)
	if e != nil {
		return -1, e
	}
	return n.(int64), nil
}

func (c *Conn) HVALS(key string) ([]interface{}, error) {
	v, e := c.Call("HVALS", key)
	if e != nil {
		return nil, e
	}
	return v.([]interface{}), nil
}

func (c *Conn) HSCAN() {}

/******************* lists commands *******************/
func (c *Conn) BLPOP(keys []string, timeout int) ([]interface{}, error) {
	args := make([]interface{}, len(keys)+1)
	for k, v := range keys {
		args[k] = v
	}
	args[len(keys)] = timeout

	v, e := c.Call("BLPOP", args...)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrKeyNotExist
	}
	return v.([]interface{}), nil
}

func (c *Conn) BRPOP(keys []string, timeout int) ([]interface{}, error) {
	args := make([]interface{}, len(keys)+1)
	for k, v := range keys {
		args[k] = v
	}
	args[len(keys)] = timeout

	v, e := c.Call("BRPOP", args...)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrKeyNotExist
	}
	return v.([]interface{}), nil
}

func (c *Conn) BRPOPLPUSH(source, dest string, timeout int) ([]byte, error) {
	v, e := c.Call("BRPOPLPUSH", source, dest, timeout)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrKeyNotExist
	}
	return v.([]byte), nil
}

func (c *Conn) LINDEX(key string, index int) ([]byte, error) {
	v, e := c.Call("LINDEX", key, index)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrKeyNotExist
	}
	return v.([]byte), nil
}

func (c *Conn) LINSERT(key, dir, pivot, value string) (int64, error) {
	if strings.ToLower(dir) != "before" && strings.ToLower(dir) != "after" {
		return -1, errors.New(CommonErrPrefix + "dir only can be (before or after)")
	}
	n, e := c.Call("LINSERT", key, dir, pivot, value)
	if e != nil {
		return -1, e
	}
	return n.(int64), nil
}

func (c *Conn) LLEN(key string) (int64, error) {
	n, e := c.Call("LLEN", key)
	if e != nil {
		return -1, e
	}
	return n.(int64), nil
}

func (c *Conn) LPOP(key string) ([]byte, error) {
	v, e := c.Call("LPOP", key)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrKeyNotExist
	}
	return v.([]byte), nil
}

func (c *Conn) LPUSH(key string, values []string) (int64, error) {
	args := make([]interface{}, len(values)+1)
	args[0] = key
	for i, v := range values {
		args[i+1] = v
	}
	n, e := c.Call("LPUSH", args...)
	if e != nil {
		return -1, e
	}
	return n.(int64), nil
}

func (c *Conn) LPUSHX(key, value string) (int64, error) {
	n, e := c.Call("LPUSHX", key, value)
	if e != nil {
		return -1, e
	}
	return n.(int64), nil
}

func (c *Conn) LRANGE(key string, start, end int) ([]interface{}, error) {
	v, e := c.Call("LRANGE", key, start, end)
	if e != nil {
		return nil, e
	}
	return v.([]interface{}), nil
}

func (c *Conn) LREM(key string, count int, value string) (int64, error) {
	n, e := c.Call("LREM", key, count, value)
	if e != nil {
		return -1, e
	}
	return n.(int64), nil
}

func (c *Conn) LSET(key string, index int, value string) ([]byte, error) {
	v, e := c.Call("LSET", key, index, value)
	if e != nil {
		return nil, e
	}
	return v.([]byte), nil
}

func (c *Conn) LTRIM(key string, start, end int) ([]byte, error) {
	v, e := c.Call("LTRIM", key, start, end)
	if e != nil {
		return nil, e
	}
	return v.([]byte), nil
}

func (c *Conn) RPOP(key string) ([]byte, error) {
	v, e := c.Call("RPOP", key)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrKeyNotExist
	}
	return v.([]byte), nil
}

func (c *Conn) RPOPLPUSH(source, dest string) ([]byte, error) {
	v, e := c.Call("RPOPLPUSH", source, dest)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, ErrKeyNotExist
	}
	return v.([]byte), nil
}

func (c *Conn) RPUSH(key string, values []string) (int64, error) {
	args := make([]interface{}, len(values)+1)
	args[0] = key
	for i, v := range values {
		args[i+1] = v
	}
	n, e := c.Call("RPUSH", args...)
	if e != nil {
		return -1, e
	}
	return n.(int64), nil
}

func (c *Conn) RPUSHX(key, value string) (int64, error) {
	n, e := c.Call("RPUSHX", key, value)
	if e != nil {
		return -1, e
	}
	return n.(int64), nil
}

/******************* sorted sets commands *******************/
func (c *Conn) ZADD()             {}
func (c *Conn) ZCARD()            {}
func (c *Conn) ZCOUNT()           {}
func (c *Conn) ZINCRBY()          {}
func (c *Conn) ZINTERSTORE()      {}
func (c *Conn) ZLEXCOUNT()        {}
func (c *Conn) ZRANGE()           {}
func (c *Conn) ZRANGEBYLEX()      {}
func (c *Conn) ZREVRANGEBYLEX()   {}
func (c *Conn) ZRANGEBYSCORE()    {}
func (c *Conn) ZRANK()            {}
func (c *Conn) ZREM()             {}
func (c *Conn) ZREMRANGEBYLEX()   {}
func (c *Conn) ZREMRANGEBYSCORE() {}
func (c *Conn) ZREVRANK()         {}
func (c *Conn) ZSCORE()           {}
func (c *Conn) ZUNIONSTORE()      {}
func (c *Conn) ZSCAN()            {}
