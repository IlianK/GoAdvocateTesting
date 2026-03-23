package main

import (
	"flag"
	"fmt"
	"os"
)

// advocate-runner - CLI to run ADVOCATE fuzzing/analysis and structure results
func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "test":
		cmdTest(os.Args[2:])
	case "compare":
		cmdCompare(os.Args[2:])
	case "-h", "--help", "help":
		usage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println(`advocate-runner - CLI to run ADVOCATE fuzzing/analysis and structure results
	Usage:
	advocate-runner test <testDir> [--recursive] [--tests-file filtered_tests.txt] [--config config.yaml]
	advocate-runner compare ...      (stub for now)

	Examples:
	advocate-runner test ./Examples/Examples_Simple/MixedDeadlock
	advocate-runner test ./Examples/Projects --recursive --tests-file ./filtered_tests.txt
	`)
	_ = flag.CommandLine
}
