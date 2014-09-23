package msgredis

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
