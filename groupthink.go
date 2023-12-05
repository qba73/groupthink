package groupthink

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/mr-joshcrane/oracle"
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
	fmt.Println("Listening on", s.Address)
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
				if item != "" {
					s.AddItem(strings.TrimSpace(item))
				}
				for _, i := range s.Items() {
					fmt.Fprintln(conn, i)
				}
				fmt.Fprintln(conn, "OK")
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
	Conn  net.Conn
	Items []string
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
	_, err := fmt.Fprintln(c.Conn, item)
	if err != nil {
		return err
	}
	c.Items = []string{}
	scanner := bufio.NewScanner(c.Conn)
	for scanner.Scan() {
		item := scanner.Text()
		if item == "OK" {
			break
		}
		c.Items = append(c.Items, item)
	}
	return nil
}

// RunAIClient creates a new GroupThink client with ChatGPT support.
func RunAIClient() {
	c, err := NewClient(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "server uri required first argument")
		os.Exit(1)
	}
	token := os.Getenv("OPENAI_API_KEY")
	if token == "" {
		fmt.Fprintln(os.Stderr, "OpenAI API token not exported")
		os.Exit(1)
	}
	o := oracle.WithGPT35Turbo()(oracle.NewOracle(token))
	// o.SetPurpose("You generate a single creative and tangential suggestion in a brainstorming session.")
	// o.GiveExample("Understand how OAuth works", "Create a CLI application that utilizes device flow")
	o.SetPurpose("Give similar suggestions.")
	o.GiveExample("rugby", "football")

	err = c.AddItem("")
	if err != nil {
		os.Exit(1)
	}

	for {
		out := c.Items
		query := strings.Join(out, "\n")
		fmt.Println("ITEMS>>>", query)
		answer, err := o.Ask(context.Background(), query)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		c.AddItem(strings.Split(answer, "\n")[0])
		fmt.Println("PRINT>>", strings.Split(answer, "\n")[0])
	}
}
