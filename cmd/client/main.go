package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/qba73/groupthink"
)

func main() {
	c, err := groupthink.NewClient(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "server uri required first argument")
		os.Exit(1)
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		err = c.AddItem(line)
		if err != nil {
			os.Exit(1)
		}
		out := c.Items
		query := strings.Join(out, "\n")
		fmt.Printf("ITEMS>>>\n%s\n", query)
	}
}
