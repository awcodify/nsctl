package main

import (
	"fmt"
	"log"
	"os"

	"nsctl/pkg/ns"
)

func main() {
	// Special case: we're being re-executed to run setup inside the namespace
	// This happens when RunWithSetup re-executes this binary with "setup-and-exec"
	if isNamespaceSetupCall() {
		handleNamespaceSetup()
		return
	}

	// Normal execution: parse user commands
	if len(os.Args) < 2 {
		showUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "run":
		handleRunCommand()
	case "ps":
		handlePsCommand()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		showUsage()
		os.Exit(1)
	}
}

// isNamespaceSetupCall checks if we're being called for internal namespace setup
func isNamespaceSetupCall() bool {
	return len(os.Args) >= 3 && os.Args[1] == "setup-and-exec"
}

// handleNamespaceSetup processes the internal namespace setup call
func handleNamespaceSetup() {
	targetCmd := os.Args[2]
	targetArgs := os.Args[3:]

	if err := ns.HandleSetupAndExec(targetCmd, targetArgs); err != nil {
		log.Fatalf("Failed to setup namespace: %v", err)
	}
}

// handleRunCommand processes the "run" command to start a container
func handleRunCommand() {
	if len(os.Args) < 3 {
		fmt.Printf("Missing command to run\n")
		fmt.Printf("Usage: %s run <command> [args...]\n", os.Args[0])
		os.Exit(1)
	}

	targetCmd := os.Args[2]
	targetArgs := os.Args[3:]

	fmt.Printf("[nsctl] Starting container with command: %s %v\n", targetCmd, targetArgs)

	// Use current executable path for re-execution
	execPath := os.Args[0]

	// Create isolated environment and run the command
	if err := ns.RunWithSetup(execPath, targetCmd, targetArgs); err != nil {
		log.Fatalf("Container failed: %v", err)
	}
}

// handlePsCommand processes the "ps" command to list containers
func handlePsCommand() {
	fmt.Printf("[nsctl] Listing containers...\n")

	containers, err := ns.ListContainers()
	if err != nil {
		log.Fatalf("Failed to list containers: %v", err)
	}

	fmt.Print(ns.FormatContainerTable(containers))
}

// showUsage displays help information
func showUsage() {
	fmt.Printf("[nsctl] Minimal Container Runtime\n\n")
	fmt.Printf("Usage:\n")
	fmt.Printf("  %s run <command> [args...]  # Run command in isolated container\n", os.Args[0])
	fmt.Printf("  %s ps                       # List running containers\n", os.Args[0])
	fmt.Printf("\nExamples:\n")
	fmt.Printf("  %s run /bin/bash           # Start isolated bash shell\n", os.Args[0])
	fmt.Printf("  %s run ls -la              # Run ls command in container\n", os.Args[0])
}
