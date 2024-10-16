package server

import (
	"net"
	"time"
)

func Listen(addr string) (*listener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &listener{l.(*net.TCPListener)}, nil
}

type listener struct {
	*net.TCPListener
}

func (ln listener) Accept() (net.Conn, error) {
	tcpconn, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	if err = tcpconn.SetKeepAlive(true); err != nil {
		return nil, err
	}
	if err = tcpconn.SetKeepAlivePeriod(3 * time.Minute); err != nil {
		return nil, err
	}
	return tcpconn, nil
}
