package main

import (
	"fmt"
	"math"
	"os"

	tl "github.com/stefanovazzocell/ToyLanguage/src"
)

const (
	ExecutionLimit = 1000000000
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: toylanguage COMMAND [OPTION]\nTry 'toylanguage help' for more information.")
		os.Exit(0)
	}
	switch os.Args[1] {
	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Usage: toylanguage run <file>\nTry 'toylanguage help' for more information.")
			os.Exit(0)
		}
		program, err := Load(os.Args[2])
		if err != nil {
			fmt.Printf("Failed to load program: %v\n", err)
			return
		}
		if program.HasExtensions(tl.ExtNet) {
			fmt.Print("Network Extension Enabled\n\n")
		}
		// Run
		if err = program.Run(ExecutionLimit); err != nil {
			fmt.Printf("\n\nProgram terminated with error: %v", err)
		}
		fmt.Println()
	case "rununlimited":
		if len(os.Args) < 3 {
			fmt.Println("Usage: toylanguage rununlimited <file>\nTry 'toylanguage help' for more information.")
			os.Exit(0)
		}
		program, err := Load(os.Args[2])
		if err != nil {
			fmt.Printf("Failed to load program: %v\n", err)
			return
		}
		if program.HasExtensions(tl.ExtNet) {
			fmt.Print("Network Extension Enabled\n\n")
		}
		// Run
		for {
			err = program.Run(math.MaxInt)
			if err == nil {
				// Terminated
				fmt.Println()
				return
			}
			if err != tl.ErrExecutionLimit {
				// Errored
				fmt.Printf("\n\nProgram terminated with error: %v", err)
				return
			}
		}
	case "help":
		displayHelp()
	default:
		fmt.Println("Usage: toylanguage COMMAND [OPTION]\nTry 'toylanguage help' for more information.")
	}
}
