package gotcpio

import (
	"context"
	"net"
	"strings"
	"sync"
)

type Emiter interface {
	Emit(string, []byte) error
}

type Server struct {
	Addr string

	listener net.Listener

	conns  []*Conn
	connMu sync.RWMutex

	// handler Handler
	onConn       func(*Conn)
	onDisconnect func(*Conn)

	room   map[string]Emiter
	roomMu sync.RWMutex
}

func NewServer(addr string) *Server {
	return &Server{
		Addr:  addr,
		room:  make(map[string]Emiter),
		conns: make([]*Conn, 0),
	}
}

func (s *Server) On(event string, handler func(*Conn)) {
	switch event {
	case "connection":
		s.onConn = handler
	case "disconnect":
		s.onDisconnect = handler
	default:
		panic("Invalid event")
	}
}

func (s *Server) handleConn(c *Conn) {
	if s.onConn != nil {
		s.onConn(c)
	}

	scan := c.newScanner()

	// Read the incoming connection
	for scan.Scan() {
		txt := scan.Text()
		mgs, data, _ := strings.Cut(txt, "\t")

		c.handleEvent(mgs, []byte(data))
	}

	if s.onDisconnect != nil { // If there is a disconnect handler
		s.onDisconnect(c)
	}

	// s.removeConn(c)
}

func (s *Server) addConn(c *Conn) {
	s.connMu.Lock()
	s.roomMu.Lock()
	defer s.connMu.Unlock()
	defer s.roomMu.Unlock()

	s.conns = append(s.conns, c)
	s.room[c.id.String()] = c
}

func (s *Server) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	s.listener = listener

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		c := NewConn(context.Background(), conn)
		s.addConn(c)

		go s.handleConn(c)
	}
}

func (s *Server) ListenWithListener(listener net.Listener) error {
	s.listener = listener

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		c := NewConn(context.Background(), conn)
		s.addConn(c)

		go s.handleConn(c)
	}
}

func (s *Server) Close() error {
	return s.listener.Close()
}
