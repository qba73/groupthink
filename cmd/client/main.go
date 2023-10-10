package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

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
	o := oracle.NewOracle(token)
	o.SetPurpose("You generate a single suggestion in a brainstorming session. Try not to repeat yourself.")
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

func Brainstorm(o *oracle.Oracle, suggestions []string) (string, error) {
	o.SetPurpose("You generate a single suggestion in a brainstorming session. Try not to repeat yourself.")
	s := strings.Join(suggestions, "\n")
	return o.Ask(context.TODO(), s)
}
