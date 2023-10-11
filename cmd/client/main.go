package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/mr-joshcrane/oracle"
	"github.com/qba73/groupthink"
)

func main() {
	c, err := groupthink.NewClient(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "server uri required first argument")
		os.Exit(1)
	}
	token := os.Getenv("OPENAI_API_KEY")
	o := oracle.WithGPT4()(oracle.NewOracle(token))
	o.SetPurpose("You generate a single creative and tangential suggestion in a brainstorming session.")
	o.GiveExample("Understand how OAuth works", "Create a CLI application that utilizes device flow")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		err = c.AddItem(line)
		if err != nil {
			os.Exit(1)
		}
		answer, err := o.Ask(context.Background(), line)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		c.AddItem(answer)
		fmt.Println(answer)
	}
}
