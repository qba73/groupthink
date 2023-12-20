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

type client chan<- string

var (
	// all incoming messages from connected clients
	messages = make(chan string)

	// new incomming connections to the server
	entering = make(chan client)

	// dropped connections
	leaving = make(chan client)
)

type Server struct {
	m     sync.RWMutex
	items []string

	Address  string
	Listener net.Listener
	lg       log.Logger

	// keep track of connected clients
	clients map[client]bool
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
	s.lg.Printf("Listening on %s", s.Address)
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
	go s.broadcast()
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			s.lg.Print(err)
			continue
		}
		//go s.handleConn(conn)
		go thinkHandler(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
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
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg) // todo: add error handling
	}
}

func thinkHandler(conn net.Conn) {
	ch := make(chan string)
	go clientWriter(conn, ch)

	clientID := conn.RemoteAddr().String()
	ch <- "Connected new client: " + clientID

	messages <- clientID + " has joined brainstorming session"
	entering <- ch

	input := bufio.NewScanner(conn)
	for input.Scan() {
		messages <- clientID + ": " + input.Text()
	}

	leaving <- ch
	messages <- clientID + " has disconnected"
	conn.Close()
}

// braodcast sends messages to connected clients
// and adds and removes clients from the pool.
func (s *Server) broadcast() {
	for {
		select {
		// broadcast message to all clients's outgoing message channels
		case msg := <-messages:
			for c := range s.clients {
				c <- msg
			}

		// a new client connects to the server:
		//  - add it to the client pool
		case c := <-entering:
			s.clients[c] = true

		// a client disconnects from the server
		//  - remove it from the pool
		//  - close the channel (client) the client uses to communicate with the server
		case cl := <-leaving:
			delete(s.clients, cl)
			close(cl)
		}
	}
}

// Start creates and starts default groupthink server.
// The server listens on a randomly assigned free port.
func Start() {
	srv := Server{
		lg:      *log.New(os.Stdout, "GROUPTHINK:", log.LstdFlags),
		items:   make([]string, 0),
		clients: make(map[client]bool),
	}
	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
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
