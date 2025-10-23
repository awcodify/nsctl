//go:build linux

package ns

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/unix"
)

// RunWithSetup creates a process with isolated namespaces and sets up the environment
// This is the main entry point for creating containers
func RunWithSetup(execPath string, command string, args []string) error {
	fmt.Printf("[ns] Creating isolated namespaces (PID, UTS, Mount)\n")
	fmt.Printf("[ns] Using executable: %s\n", execPath)

	// Re-execute ourselves with special arguments to run setup inside the namespace
	// This two-step process is necessary because namespace setup must happen inside the namespace
	setupArgs := []string{"setup-and-exec", command}
	setupArgs = append(setupArgs, args...)

	cmd := exec.Command(execPath, setupArgs...)

	// Configure namespace isolation using clone flags
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: unix.CLONE_NEWUTS | // Isolate hostname/domainname
			unix.CLONE_NEWPID | // Isolate process IDs (new PID namespace)
			unix.CLONE_NEWNS, // Isolate filesystem mounts
	}

	// Connect container I/O to parent terminal
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start the namespaced process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start namespace process: %v", err)
	}

	containerPID := cmd.Process.Pid
	fmt.Printf("[ns] Container started with PID %d\n", containerPID)

	// Wait for container to finish and return its exit status
	return cmd.Wait()
}

// HandleSetupAndExec runs inside the new namespace to set up the environment
// and then execute the target command
func HandleSetupAndExec(targetCmd string, targetArgs []string) error {
	fmt.Printf("[ns] Setting up isolated environment...\n")

	// Step 1: Set custom hostname in the UTS namespace
	newHostname := "container"
	fmt.Printf("[ns] Setting hostname to '%s'\n", newHostname)
	if err := unix.Sethostname([]byte(newHostname)); err != nil {
		return fmt.Errorf("failed to set hostname: %v", err)
	}

	// Step 2: Mount /proc for the new PID namespace
	// This gives us the isolated view of processes (ps, top, etc. will work correctly)
	fmt.Printf("[ns] Mounting /proc filesystem for isolated process view\n")
	if err := unix.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		return fmt.Errorf("failed to mount /proc: %v", err)
	}

	// Step 3: Execute the target command
	fmt.Printf("[ns] Executing target command: %s %v\n", targetCmd, targetArgs)

	// Find the full path to the command
	targetPath, err := exec.LookPath(targetCmd)
	if err != nil {
		return fmt.Errorf("command not found: %s (%v)", targetCmd, err)
	}

	// Replace the current process with the target command
	// This makes the target command PID 1 in the new namespace
	execArgs := append([]string{targetCmd}, targetArgs...)

	fmt.Printf("[ns] Replacing process with target command...\n")
	return syscall.Exec(targetPath, execArgs, os.Environ())
}

// Legacy function kept for compatibility - prefer RunWithSetup
func Run(command string, args []string) error {
	fmt.Printf("[ns] Using legacy Run function - consider using RunWithSetup\n")

	cmd := exec.Command(command, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command in namespace: %v", err)
	}

	containerPID := cmd.Process.Pid
	fmt.Printf("[ns] Started process with PID %d\n", containerPID)

	return cmd.Wait()
}
