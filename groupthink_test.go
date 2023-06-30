package groupthink_test

import (
	"fmt"
	"io"
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/qba73/groupthink"
)

func TestReceivingHelloMessage(t *testing.T) {

	server, err := groupthink.Start()
	if err != nil {
		t.Fatal(err)
	}

	conn, err := net.Dial("tcp", server.Address)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Fprintln(conn, "Hello!")

	data, err := io.ReadAll(conn)
	if err != nil {
		t.Fatal(err)
	}

	want := []byte("Hello!\n")

	if !cmp.Equal(want, data) {
		t.Error(cmp.Diff(string(want), string(data)))
	}

}

func TestSendItems(t *testing.T) {

	server, err := groupthink.Start()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%p\n", server)

	conn, err := net.Dial("tcp", server.Address)
	if err != nil {
		t.Fatal(err)
	}

	got := server.Items()
	if got != nil {
		t.Errorf("want nil, got: %v", got)
	}

	fmt.Fprintln(conn, "item1")

	got = server.Items()
	want := []string{"item1\n"}

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}
