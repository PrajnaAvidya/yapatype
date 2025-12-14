package main

import (
	"os"

	"github.com/PrajnaAvidya/yapatype/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
