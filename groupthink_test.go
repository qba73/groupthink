package groupthink_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/qba73/groupthink"
)

func TestServerStoresItemSentByClient(t *testing.T) {

	srv := groupthink.Server{}
	err := srv.Listen(":0")
	if err != nil {
		t.Fatal(err)
	}
	go srv.Serve()

	conn, err := net.Dial("tcp", srv.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	fmt.Fprintln(conn, "Hello")

	var dummy string
	_, err = fmt.Fscanln(conn, &dummy)
	if err != nil {
		t.Fatal(err)
	}

	got := srv.Items()
	want := []string{"Hello"}

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}

}

// func TestSendItems(t *testing.T) {

// 	server, err := groupthink.Start()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	fmt.Printf("%p\n", server)

// 	conn, err := net.Dial("tcp", server.Address)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	got := server.Items()
// 	if got != nil {
// 		t.Errorf("want nil, got: %v", got)
// 	}

// 	fmt.Fprintln(conn, "item1")

// 	got = server.Items()
// 	want := []string{"item1\n"}

// 	if !cmp.Equal(want, got) {
// 		t.Error(cmp.Diff(want, got))
// 	}
// }

// func TestStartServer(t *testing.T) {
// 	server := groupthink.Server{
// 		Address: "localhost:8087",
// 	}

// 	if err := server.Start(); err != nil {
// 		t.Fatal()
// 	}

// 	conn := waitForConn("locahost:8087")

// 	fmt.

// }

// waitForConn returns connection to the server.
//
// If the server is not ready yet to accept connections,
// it waits 10ms before trying to connect again. As we launch
// the server in a separate goroutine, waiting
// until the server is ready to accept connections is necessary.
func waitForConn(addr string) net.Conn {
	for {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			return conn
		}
		time.Sleep(10 * time.Millisecond)
	}
}
