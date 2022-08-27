package main

import (
	"fmt"
	"os"

	tl "github.com/stefanovazzocell/ToyLanguage/src"
)

// Load a given program
func Load(programSrc string) (tl.Program, error) {
	// Open source
	file, err := os.Open(programSrc)
	if err != nil {
		fmt.Printf("Error opening %p: %v\n", &programSrc, err)
	}
	defer file.Close()
	// Parse
	prog, err := tl.NewProgram(file)
	return prog, err
}

// Outputs a help guide in the screen and quits
func displayHelp() {
	fmt.Print(`ToyLanguage Help

run <file>          - Run a program
rununlimited <file> - Run a program with no execution limits 
help                - Display this guide
`)
	os.Exit(0)
}
