package main

import (
	"github.com/cwind-cli/cwind/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
