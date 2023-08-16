package groupthink_test

import (
	"bufio"
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

	fmt.Fprintln(conn, "ADD Hello")

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

	fmt.Fprintln(conn1, "ADD First Idea")

	var dummy string
	_, err = fmt.Fscanln(conn1, &dummy)
	if err != nil {
		t.Fatal(err)
	}

	conn2, err := net.Dial("tcp", srv.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	fmt.Fprintln(conn2, "ADD Second Idea")

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

func TestServerStoresItem(t *testing.T) {
	t.Parallel()

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

	fmt.Fprintln(conn, "ADD new item")

	var item string
	_, err = fmt.Fscanln(conn, &item)
	if err != nil {
		t.Fatal(err)
	}

	got := srv.Items()
	want := []string{"new item"}

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestServerRespondsWithListOfItems(t *testing.T) {
	t.Parallel()

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

	fmt.Fprintln(conn, "ADD new item")

	var item string
	_, err = fmt.Fscanln(conn, &item)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Fprintln(conn, "LIST")

	scanner := bufio.NewScanner(conn)

	var items []string

	for scanner.Scan() {
		item := scanner.Text()
		if item == "Thanks" {
			break
		}
		items = append(items, item)
	}

	want := []string{"new item"}

	if !cmp.Equal(want, items) {
		t.Error(cmp.Diff(want, items))
	}

}
