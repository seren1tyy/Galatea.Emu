package network

import (
	"net"
)

// ListenAndServe запускает TCP-сервер на указанном адресе
func ListenAndServe(addr string, handler func(net.Conn)) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go handler(conn)
	}
}
