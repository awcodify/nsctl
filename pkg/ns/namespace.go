//go:build linux

package ns

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/unix"
)

// Run creates a new process with isolated UTS, PID, and mount namespaces,
// mounts /proc, sets hostname, and executes the specified command
func Run(command string, args []string) error {
	fmt.Printf("[ns] creating PID, UTS, and mount namespaces\n")

	// Create the command that will run in the new namespaces
	cmd := exec.Command(command, args...)

	// Set up namespace isolation using Clone flags
	// CLONE_NEWUTS: isolate hostname and domainname
	// CLONE_NEWPID: isolate process IDs (new PID namespace)
	// CLONE_NEWNS: isolate mount points (mount namespace)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}

	// Connect stdin, stdout, stderr to the parent process
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start the process in the new namespaces
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command in namespace: %v", err)
	}

	// Get the PID of the new process for logging
	containerPID := cmd.Process.Pid
	fmt.Printf("[ns] started container process with PID %d\n", containerPID)

	// Wait for the process to complete
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("container process exited with error: %v", err)
	}

	return nil
}

// setupNamespaceEnvironment is called from within the new namespace
// to set up the isolated environment (mount /proc, set hostname)
func setupNamespaceEnvironment() error {
	fmt.Printf("[ns] setting up namespace environment\n")

	// Set hostname to "container" in the new UTS namespace
	fmt.Printf("[ns] setting hostname to 'container'\n")
	if err := unix.Sethostname([]byte("container")); err != nil {
		return fmt.Errorf("failed to set hostname: %v", err)
	}

	// Mount /proc inside the new PID namespace so commands like ps work correctly
	// This gives us the isolated view of processes in the new PID namespace
	fmt.Printf("[ns] mounting /proc filesystem\n")
	if err := unix.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		return fmt.Errorf("failed to mount /proc: %v", err)
	}

	return nil
}

// RunWithSetup creates a process with namespaces and runs setup inside it
func RunWithSetup(execPath string, command string, args []string) error {
	fmt.Printf("[ns] creating PID, UTS, and mount namespaces with internal setup\n")

	// Use the provided executable path (obtained from the parent process)
	// This avoids the /proc/self/exe issue inside the new mount namespace
	fmt.Printf("[ns] using executable path: %s\n", execPath)

	// Create arguments for re-executing ourselves with the setup-and-exec command
	wrapperArgs := []string{execPath, "setup-and-exec", command}
	wrapperArgs = append(wrapperArgs, args...)

	cmd := exec.Command(execPath, wrapperArgs[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: unix.CLONE_NEWUTS | unix.CLONE_NEWPID | unix.CLONE_NEWNS,
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start namespace process: %v", err)
	}

	containerPID := cmd.Process.Pid
	fmt.Printf("[ns] started container process with PID %d\n", containerPID)

	return cmd.Wait()
}

// HandleSetupAndExec is called when the program is re-executed with "setup-and-exec"
// This allows us to run setup code inside the new namespace
func HandleSetupAndExec(targetCmd string, targetArgs []string) error {
	// We're now inside the new namespace, set up the environment
	if err := setupNamespaceEnvironment(); err != nil {
		log.Fatalf("Failed to setup namespace environment: %v", err)
	}

	// Execute the target command
	fmt.Printf("[ns] executing target command: %s %v\n", targetCmd, targetArgs)

	// Replace current process with the target command
	// This ensures the target command becomes PID 1 in the new namespace
	targetPath, err := exec.LookPath(targetCmd)
	if err != nil {
		return fmt.Errorf("failed to find command %s: %v", targetCmd, err)
	}

	// Prepare arguments (argv[0] should be the command name)
	execArgs := append([]string{targetCmd}, targetArgs...)

	// Execute the command, replacing the current process
	return syscall.Exec(targetPath, execArgs, os.Environ())
}
