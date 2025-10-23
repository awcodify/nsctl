package main

import (
	"fmt"
	"log"
	"os"

	"nsctl/pkg/ns"
)

func main() {
	// Handle the special case where we're re-executing ourselves
	// for namespace setup (called from RunWithSetup)
	if len(os.Args) >= 3 && os.Args[1] == "setup-and-exec" {
		targetCmd := os.Args[2]
		targetArgs := os.Args[3:]

		if err := ns.HandleSetupAndExec(targetCmd, targetArgs); err != nil {
			log.Fatalf("Failed to setup and exec: %v", err)
		}
		return
	}

	// Parse command line arguments
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s run <command> [args...]\n", os.Args[0])
		fmt.Printf("       %s ps\n", os.Args[0])
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "run":
		if len(os.Args) < 3 {
			fmt.Printf("Usage: %s run <command> [args...]\n", os.Args[0])
			os.Exit(1)
		}

		// Extract the command and its arguments
		targetCmd := os.Args[2]
		targetArgs := os.Args[3:]

		fmt.Printf("[nsctl] starting container with command: %s %v\n", targetCmd, targetArgs)

		// Use os.Args[0] as the executable path - this is more reliable than os.Executable()
		// especially when /proc/self/exe might not be available
		execPath := os.Args[0]

		// Run the command in isolated namespaces
		if err := ns.RunWithSetup(execPath, targetCmd, targetArgs); err != nil {
			log.Fatalf("Failed to run container: %v", err)
		}

	case "ps":
		fmt.Printf("[nsctl] listing containers (not implemented yet)\n")

	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Printf("Available commands: run, ps\n")
		os.Exit(1)
	}
}
