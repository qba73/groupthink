package groupthink_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/qba73/groupthink"
)

func TestStoreDoesNotStoreDublicatedItems(t *testing.T) {
	t.Parallel()

	store := &groupthink.Store{
		Ideas: make(map[string]bool),
	}

	store.Add("Rugby")
	store.Add("Tennis")
	store.Add("Rugby")

	want := []string{"Rugby", "Tennis"}
	got := store.List()

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got, cmpopts.SortSlices(func(i, j string) bool {
			return i > j
		})))
	}
}

func TestServerStoresItemSentByClient(t *testing.T) {
	srv := groupthink.NewServer()
	err := srv.Listen(":0")
	if err != nil {
		t.Fatal(err)
	}
	go srv.Serve()

	client, err := groupthink.NewClient(srv.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	err = client.AddItem("Hello")
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
	srv := groupthink.NewServer()
	err := srv.Listen(":0")
	if err != nil {
		t.Fatal(err)
	}
	go srv.Serve()

	client1, err := groupthink.NewClient(srv.Address)
	if err != nil {
		t.Fatal(err)
	}
	err = client1.AddItem("First Idea")
	if err != nil {
		t.Fatal(err)
	}

	client2, err := groupthink.NewClient(srv.Address)
	if err != nil {
		t.Fatal(err)
	}

	err = client2.AddItem("Second Idea")
	if err != nil {
		t.Fatal(err)
	}

	got := srv.Items()
	want := []string{"First Idea", "Second Idea"}

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got, cmpopts.SortSlices(func(i, j string) bool {
			return i > j
		})))
	}
}

func TestServerRespondsWithListOfItems(t *testing.T) {
	t.Parallel()

	srv := groupthink.NewServer()
	err := srv.Listen(":0")
	if err != nil {
		t.Fatal(err)
	}

	go srv.Serve()

	client, err := groupthink.NewClient(srv.Address)
	if err != nil {
		t.Fatal(err)
	}

	err = client.AddItem("First idea")
	if err != nil {
		t.Fatal(err)
	}
	err = client.AddItem("Second idea")
	if err != nil {
		t.Fatal(err)
	}
}
