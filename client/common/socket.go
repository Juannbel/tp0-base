package common

import (
	"net"
)

// Implements a TCP socket connection.
type Socket struct {
	conn net.Conn
}

func Connect(address string) (*Socket, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &Socket{conn: conn}, nil
}

func (s *Socket) Close() error {
	return s.conn.Close()
}

func (s *Socket) SendAll(data []byte) error {
	totalSent := 0
	for totalSent < len(data) {
		n, err := s.conn.Write(data[totalSent:])
		if err != nil {
			return err
		}
		totalSent += n
	}
	return nil
}

func (s *Socket) ReceiveAll(len int) ([]byte, error) {
	buf := make([]byte, len)
	totalReceived := 0
	for totalReceived < len {
		n, err := s.conn.Read(buf[totalReceived:])
		if err != nil {
			return nil, err
		}
		totalReceived += n
	}
	return buf, nil
}