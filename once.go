package goredis

import (
	"errors"
)

type OnceConn struct {
	*Conn
}

func NewOnceConn() *OnceConn {
	return &OnceConn{}
}

func (mp *MultiPool) CallOnce(address string) *OnceConn {
	oc := NewOnceConn()
	c := mp.PopByAddr(address)
	if c == nil {
		oc.Conn = NewConn(nil, 0, 0, 0, false, nil, "")
		oc.Conn.err = errors.New("get a nil conn address=" + address)
		return oc
	}
	oc.Conn = c
	oc.Conn.isOnce = true
	return oc
}
