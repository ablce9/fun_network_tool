package main

import (
	"golang.org/x/net/websocket"
	"container/list"
)

// WSServer ...
type WSServer struct {
	wsConnPool *list.List
}

// LenWSPool shows a length of current pool size.
func (s *WSServer) LenWSPool() int {
	return s.wsConnPool.Len()
}

// MaxWSConn defines max connection numbers.
const MaxWSConn = 256

// PullPeer ...
func (s *WSServer) PullPeer(c *list.Element) {
	s.wsConnPool.Remove(c)
}

// AddPeer adds a new peer to WSPool
func (s *WSServer) AddPeer(Conn *websocket.Conn) *list.Element {
	var element *list.Element
	l := s.LenWSPool()
	if l >= MaxWSConn {
		return nil
	}
	element = s.wsConnPool.PushFront(Conn)
	return element
}

// GetPool ...
func (s *WSServer) GetPool() *list.List {
	return s.wsConnPool
}
