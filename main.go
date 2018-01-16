package main

import (
	"fmt"
	"os"

	flag "github.com/ogier/pflag"
)

// flags
var (
	fileName     string
	startingLine string
)

func main() {
	flag.Parse()

	// if user does not supply flags, print usage
	// we can clean this up later by putting this into its own function
	if flag.NFlag() == 0 {
		fmt.Printf("Usage: %s [options]\n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Printf("Searching file(s): %s\n", fileName)
	fmt.Printf("At line: %s\n", startingLine)
}

func init() {
	flag.StringVarP(&fileName, "fileName", "f", "", "file name")
	flag.StringVarP(&startingLine, "startingLine", "l", "1", "starting line")
}
