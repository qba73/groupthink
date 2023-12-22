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
	"slices"
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

type Storer interface {
	Add(string)
	List() []string
}

type Store struct {
	mu    sync.Mutex
	Ideas map[string]bool
}

func (s *Store) Add(idea string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	idea = strings.TrimSpace(idea)
	if _, ok := s.Ideas[idea]; !ok {
		s.Ideas[idea] = true
	}
}

func (s *Store) List() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	ix := make([]string, 0, len(s.Ideas))
	for k := range s.Ideas {
		ix = append(ix, k)
	}
	return ix
}

type Server struct {
	Address   string
	Listener  net.Listener
	ErrLogger log.Logger

	inShutdown atomic.Bool // set to true when server is in shutdown

	// storage for brainstorming ideas
	store Storer
}

func NewServer() *Server {
	return &Server{
		store: &Store{
			Ideas: make(map[string]bool),
		},
	}
}

func (srv *Server) Close() error {
	srv.inShutdown.Store(true)
	// todo close connections (?)
	return nil
}

// Items returns all stored items.
func (srv *Server) Items() []string {
	ix := srv.store.List()
	slices.Sort(ix)
	return ix
}

// AddItem takes a string representing the item name and stores it in the store.
func (srv *Server) AddItem(i string) {
	srv.store.Add(i)
}

func (srv *Server) Listen(addr string) error {
	if addr == "" {
		addr = ":0"
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	srv.Listener = l
	srv.Address = l.Addr().String()
	return nil
}

func (srv *Server) Serve() {
	go broadcast()
	for {
		conn, err := srv.Listener.Accept()
		if err != nil {
			srv.ErrLogger.Print(err)
			continue
		}
		go srv.thinkHandler(conn)
	}
}

func (srv *Server) ListenAndServe() error {
	if err := srv.Listen(srv.Address); err != nil {
		return err
	}
	srv.Serve()
	return nil
}

// func (srv *Server) handleConn(conn net.Conn) {
// 	defer conn.Close()
// 	scanner := bufio.NewScanner(conn)
// 	for scanner.Scan() {
// 		item := scanner.Text()
// 		if item != "" {
// 			srv.AddItem(strings.TrimSpace(item))
// 		}
// 		for _, i := range srv.Items() {
// 			fmt.Fprintln(conn, i)
// 		}
// 		fmt.Fprintln(conn, "OK")
// 	}
// }

// braodcast sends messages to connected clients
// and adds and removes clients from the pool.
func broadcast() {
	clients := make(map[client]bool)
	for {
		select {
		// broadcast message to all clients's outgoing message channels
		case msg := <-messages:
			for c := range clients {
				c <- msg
			}

		// a new client connects to the server:
		//  - add it to the client pool
		case c := <-entering:
			clients[c] = true

		// a client disconnects from the server
		//  - remove it from the pool
		//  - close the channel (client) the client uses to communicate with the server
		case cl := <-leaving:
			delete(clients, cl)
			close(cl)
		}
	}
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg) // todo: add error handling
	}
}

func (srv *Server) thinkHandler(conn net.Conn) {
	defer conn.Close()

	ch := make(chan string)
	go clientWriter(conn, ch)

	clientID := conn.RemoteAddr().String()
	ch <- "Connected new client: " + clientID

	messages <- clientID + " has joined brainstorming session"
	entering <- ch

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		item := scanner.Text()
		item = strings.TrimSpace(item)
		if item != "" {
			srv.AddItem(item)
		}
		messages <- clientID + ": " + item

		ix := srv.store.List()
		slices.Sort(ix)
		for _, i := range ix {
			messages <- i
		}
		// signal to client to disconnect
		ch <- "OK"
	}

	leaving <- ch
	messages <- clientID + " has disconnected"
}

// Start creates and starts a groupthink server.
func Start() {
	srv := NewServer()
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
