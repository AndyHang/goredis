package msgredis

// all return integer is int64
import (
	"errors"
	"fmt"
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
	fmt.Println(string(r))
	return r, nil
}

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

func (c *Conn) SET(key, value string) ([]byte, error) {
	v, e := c.Call("SET", key, value)
	if e != nil {
		return nil, e
	}
	return v.([]byte), nil
}

func (c *Conn) GET(key string) ([]byte, error) {
	v, e := c.Call("GET", key)
	if e != nil {
		return nil, e
	}
	return v.([]byte), nil
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
	return n.([]byte), nil
}

func (c *Conn) SRANDMEMBER(key string, count int) ([][]byte, error) {
	if count == 0 {
		v, e := c.Call("SRANDMEMBER", key)
		if e != nil {
			return nil, e
		}
		members := make([][]byte, 1)
		members[0] = n.([]byte)
		return members, nil
	}
	v, e := c.Call("SRANDMEMBER", key)
	if e != nil {
		return nil, e
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

func (c *Conn) SSCAN(key string, cursor int)
