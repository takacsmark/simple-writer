package main

import (
	"fmt"
	"os"

	"simple-writer/internal/editor"
)

func main() {
	if err := editor.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
