package main

import (
	"os"

	"github.com/dunkinfrunkin/kit/cmd/kit/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
