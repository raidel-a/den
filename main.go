package main

import (
	"den/internal/cli"
	"fmt"
	"os"
)

func main() {
	app := cli.New()
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
