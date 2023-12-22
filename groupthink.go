package groupthink

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/mr-joshcrane/oracle"
)

var ErrServerClosed = errors.New("groupthink: Server closed")

type client chan<- string

var (
	messages = make(chan string) // all incoming messages from connected clients
	entering = make(chan client) // new incomming connections to the server
	leaving  = make(chan client) // dropped connections
)

type Server struct {
	Address   string
	Listener  net.Listener
	ErrLogger log.Logger

	inShutdown atomic.Bool // set to true when server is in shutdown

	mu      sync.Mutex
	items   []string
	clients map[client]bool // keep track of connected clients
}

func NewServer() *Server {
	return &Server{
		clients: make(map[client]bool),
	}
}

func (s *Server) Close() error {
	s.inShutdown.Store(true)
	return nil
}

// Items returns all stored items.
func (s *Server) Items() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.items
}

// AddItem takes a string representing the item name and stores it in the store.
func (s *Server) AddItem(i string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, i)
}

func (s *Server) Listen(addr string) error {
	if addr == "" {
		addr = ":0"
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.Listener = l
	s.Address = l.Addr().String()
	return nil
}

func (s *Server) Serve() {
	//go srv.broadcast()
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			s.ErrLogger.Print(err)
			continue
		}
		go s.handleConn(conn)
		//go thinkHandler(conn)
	}
}

func (s *Server) ListenAndServe() error {
	if err := s.Listen(s.Address); err != nil {
		return err
	}
	s.Serve()
	return nil
}

func (srv *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		item := scanner.Text()
		if item != "" {
			srv.AddItem(strings.TrimSpace(item))
		}
		for _, i := range srv.Items() {
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
	defer conn.Close()

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

// Start creates and starts a groupthink server.
func Start() {
	srv := Server{
		ErrLogger: *log.New(os.Stderr, "GROUPTHINK:", log.LstdFlags),
		items:     make([]string, 0),
		clients:   make(map[client]bool),
	}
	if err := srv.Listen(":0"); err != nil {
		panic(err)
	}
	fmt.Fprintln(os.Stdout, "listening on: "+srv.Address)
	srv.Serve()
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

func RunClient(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}

	done := make(chan bool)
	go func() {
		io.Copy(os.Stdout, conn)
		done <- true
	}()

	if _, err := io.Copy(conn, os.Stdin); err != nil {
		panic(err)
	}
	conn.Close()
	<-done
}
