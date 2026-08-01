package main

import (
	"flag"
	"fmt"
	"os"
)

var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("remy %s\n", version)
		os.Exit(0)
	}

	fmt.Println("Remy starting...")
}
