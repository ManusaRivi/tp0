package network

import (
	"net"
)

type Socket struct {
	conn net.Conn
}

func NewSocket(conn net.Conn) *Socket {
	socket := &Socket{
		conn: conn,
	}
	return socket
}

func (s *Socket) Send_all(data []byte) error {
	bytesSent := 0
	for bytesSent < len(data) {
		n, err := s.conn.Write(data[bytesSent:])
		if err != nil {
			return err
		}
		bytesSent += n
	}
	return nil
}

func (s *Socket) Receive_all(data []byte) error {
	bytesReceived := 0
	for bytesReceived < len(data) {
		n, err := s.conn.Read(data[bytesReceived:])
		if err != nil {
			return err
		}
		bytesReceived += n
	}
	return nil
}

func (s *Socket) Close() error {
	return s.conn.Close()
}
