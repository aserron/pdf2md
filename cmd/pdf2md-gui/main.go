package main

import (
	"fmt"
	"os"
)

var version = "0.1.0-dev"

func main() {
	fmt.Fprintf(os.Stdout, "Starting pdf2md Windows GUI (v%s)...\n", version)
}
