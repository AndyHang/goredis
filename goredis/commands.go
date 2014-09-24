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

func (c *Conn) SET(key, value string) {

}

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
