package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

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
	o := oracle.WithGPT35Turbo()(oracle.NewOracle(token))
	// o.SetPurpose("You generate a single creative and tangential suggestion in a brainstorming session.")
	// o.GiveExample("Understand how OAuth works", "Create a CLI application that utilizes device flow")
	o.SetPurpose("Give similar suggestions.")
	o.GiveExample("rugby", "football")

	err = c.AddItem("")
	if err != nil {
		os.Exit(1)
	}

	for {
		out := c.Items
		query := strings.Join(out, "\n")
		fmt.Println("ITEMS>>>", query)

		answer, err := o.Ask(context.Background(), query)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		r := rand.Intn(20)
		time.Sleep(time.Duration(r) * time.Second)

		c.AddItem(strings.Split(answer, "\n")[0])
		fmt.Println("PRINT>>", strings.Split(answer, "\n")[0])

	}
}
