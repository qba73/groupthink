package groupthink_test

import (
	"fmt"
	"net"
	"testing"

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

func TestServerStoresItemsSentByMultipleClients(t *testing.T) {
	srv := groupthink.Server{}
	err := srv.Listen(":0")
	if err != nil {
		t.Fatal(err)
	}
	go srv.Serve()

	conn1, err := net.Dial("tcp", srv.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	fmt.Fprintln(conn1, "First Idea")

	var dummy string
	_, err = fmt.Fscanln(conn1, &dummy)
	if err != nil {
		t.Fatal(err)
	}

	conn2, err := net.Dial("tcp", srv.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	fmt.Fprintln(conn2, "Second Idea")

	_, err = fmt.Fscanln(conn2, &dummy)
	if err != nil {
		t.Fatal(err)
	}

	got := srv.Items()
	want := []string{"First Idea", "Second Idea"}

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}
