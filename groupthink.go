package groupthink

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

type Server struct {
	Address string
	items   []string
	m       sync.RWMutex
}

func (s *Server) Items() []string {
	s.m.RLock()
	defer s.m.RUnlock()
	return s.items
}

func (s *Server) AddItem(i string) {
	s.m.Lock()
	defer s.m.Unlock()
	s.items = append(s.items, i)
}

func Start() (*Server, error) {
	s := &Server{
		Address: "localhost:8083",
	}
	l, err := net.Listen("tcp", s.Address)
	if err != nil {
		return nil, err
	}

	go func() {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		fmt.Fprintln(conn, "Hello!")

		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			item := scanner.Text()
			s.AddItem(item)
		}

		conn.Close()
	}()

	return s, nil
}
