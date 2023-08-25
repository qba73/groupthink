package groupthink

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
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
	l, err := net.Listen("tcp", s.Address)
	if err != nil {
		return err
	}
	s.Listener = l
	s.Address = l.Addr().String()
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

			scanner := bufio.NewScanner(conn)
			for scanner.Scan() {

				item := scanner.Text()
				if strings.HasPrefix(item, "ADD") {
					s.AddItem(strings.TrimSpace(strings.TrimPrefix(item, "ADD")))
					fmt.Fprintf(conn, "OK\n")
				}

				if strings.HasPrefix(item, "LIST") {
					for _, i := range s.Items() {
						fmt.Fprintln(conn, i)
					}
					fmt.Fprintf(conn, "OK\n")
				}
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

type Client struct {
	Conn net.Conn
}

func NewClient(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Client{
		Conn: conn,
	}, nil
}

func (c *Client) AddItem(item string) error {
	_, err := fmt.Fprintf(c.Conn, "ADD %s\n", item)
	if err != nil {
		return err
	}
	var res string
	_, err = fmt.Fscanln(c.Conn, &res)
	if err != nil {
		return err
	}
	if res != "OK" {
		return fmt.Errorf("unexpected response: %s", res)
	}
	return nil
}

func (c *Client) ListItems() ([]string, error) {
	fmt.Fprintln(c.Conn, "LIST")
	scanner := bufio.NewScanner(c.Conn)
	var items []string
	for scanner.Scan() {
		item := scanner.Text()
		if item == "OK" {
			break
		}
		items = append(items, item)
	}
	return items, scanner.Err()
}
