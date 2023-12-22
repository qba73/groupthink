package main

import (
	"os"

	"github.com/qba73/groupthink"
)

func main() {
	groupthink.RunClient(os.Args[1])
}
