package main

import (
	"fmt"
	"log"
	"os"

	"github.com/akamensky/argparse"
)

func main() {
	log.SetFlags(log.Flags() | log.Lmicroseconds)

	parser := argparse.NewParser("testcontainers", "Test Containers App")

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		return
	}
}
