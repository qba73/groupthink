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
