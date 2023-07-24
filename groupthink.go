package groupthink

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
)

type Server struct {
	Address  string
	Listener net.Listener
	items    []string
	m        sync.RWMutex
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

func (s *Server) ListenAndServe() error {
	if err := s.Listen(s.Address); err != nil {
		return err
	}
	s.Serve()
	return nil
}

func (s *Server) Listen(addr string) error {
	s.Address = addr
	l, err := net.Listen("tcp", s.Address)
	if err != nil {
		return err
	}
	s.Listener = l
	return nil
}

func (s *Server) Serve() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go func() {
			defer conn.Close()
			fmt.Fprintln(conn, "Hello!")

			scanner := bufio.NewScanner(conn)
			for scanner.Scan() {
				item := scanner.Text()
				s.AddItem(item)
				fmt.Fprintf(conn, "Thanks")
			}
		}()
	}
}

func Start() error {
	s := &Server{
		Address: "localhost:8083",
	}
	err := s.Listen(s.Address)
	if err != nil {
		return err
	}
	s.Serve()
	return nil
}
