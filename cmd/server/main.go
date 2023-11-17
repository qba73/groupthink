package main

import (
	"github.com/qba73/groupthink"
)

func main() {
	srv := groupthink.Server{}
	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
