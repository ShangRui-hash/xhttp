package testutils

import "net"

// TCPServer creates a new tcp server that returns a response
type TCPServer struct {
	URL      string
	listener net.Listener
}

// NewTCPServer creates a new TCP server from a handler
func NewTCPServer(handler func(conn net.Conn)) *TCPServer {
	server := &TCPServer{}

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	server.URL = l.Addr().String()
	server.listener = l

	go func() {
		for {
			// Listen for an incoming connection.
			conn, err := l.Accept()
			if err != nil {
				continue
			}
			// Handle connections in a new goroutine.
			go handler(conn)
		}
	}()
	return server
}

// Close closes the TCP server
func (s *TCPServer) Close() {
	s.listener.Close()
}
