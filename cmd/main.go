package main

import (
	"os"
	"github.com/tobias-urdin/snapback/internal/command"
)

func main() {
	cmd := command.New()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
