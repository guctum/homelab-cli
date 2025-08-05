package main

import (
	"os"

	"grognet.dev/homelab-cli-build/pkg/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}