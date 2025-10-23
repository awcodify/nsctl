//go:build linux

package ns

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// ContainerInfo holds information about a running container
type ContainerInfo struct {
	ID        string    `json:"id"`
	PID       int       `json:"pid"`
	Command   string    `json:"command"`
	Args      []string  `json:"args"`
	StartTime time.Time `json:"start_time"`
	Status    string    `json:"status"`
}

const (
	// Standard Linux runtime directory for container metadata
	// Following Filesystem Hierarchy Standard (FHS) - cleared on reboot
	defaultStateDir  = "/var/run/nsctl"
	containerFileExt = ".json"
)

var (
	// Current state directory (can be changed at runtime for fallback)
	currentStateDir = defaultStateDir
)

// ensureStateDir creates the state directory if it doesn't exist
// Uses /var/run/nsctl (standard location) with fallback to user directory if no permissions
func ensureStateDir() error {
	// Try to create the standard system directory first
	if err := os.MkdirAll(currentStateDir, 0755); err != nil {
		// If we can't write to /var/run (permission denied), use user fallback
		if os.IsPermission(err) {
			fmt.Printf("[ns] Permission denied for %s, using user directory fallback\n", currentStateDir)
			userStateDir := filepath.Join(os.Getenv("HOME"), ".nsctl", "run")
			if fallbackErr := os.MkdirAll(userStateDir, 0755); fallbackErr != nil {
				return fmt.Errorf("failed to create state directory: %v (fallback failed: %v)", err, fallbackErr)
			}
			// Update to use the fallback directory
			currentStateDir = userStateDir
			fmt.Printf("[ns] Using fallback state directory: %s\n", currentStateDir)
			return nil
		}
		return fmt.Errorf("failed to create state directory %s: %v", currentStateDir, err)
	}

	fmt.Printf("[ns] Using state directory: %s\n", currentStateDir)
	return nil
}

// generateContainerID creates a simple container ID based on timestamp and PID
func generateContainerID(pid int) string {
	return fmt.Sprintf("nsctl_%d_%d", time.Now().Unix(), pid)
}

// getContainerFilePath returns the path to a container's metadata file
func getContainerFilePath(containerID string) string {
	return filepath.Join(currentStateDir, containerID+containerFileExt)
}

// RegisterContainer saves container information to persistent storage
func RegisterContainer(pid int, command string, args []string) (string, error) {
	if err := ensureStateDir(); err != nil {
		return "", err
	}

	containerID := generateContainerID(pid)

	containerInfo := ContainerInfo{
		ID:        containerID,
		PID:       pid,
		Command:   command,
		Args:      args,
		StartTime: time.Now(),
		Status:    "running",
	}

	// Save container info to JSON file
	data, err := json.MarshalIndent(containerInfo, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal container info: %v", err)
	}

	filePath := getContainerFilePath(containerID)
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write container info: %v", err)
	}

	fmt.Printf("[ns] Registered container %s with PID %d\n", containerID, pid)
	return containerID, nil
}

// UnregisterContainer removes container information when it stops
func UnregisterContainer(containerID string) error {
	filePath := getContainerFilePath(containerID)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove container info: %v", err)
	}
	fmt.Printf("[ns] Unregistered container %s\n", containerID)
	return nil
}

// ListContainers returns information about all tracked containers
func ListContainers() ([]ContainerInfo, error) {
	if err := ensureStateDir(); err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(currentStateDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read state directory: %v", err)
	}

	var containers []ContainerInfo

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), containerFileExt) {
			continue
		}

		filePath := filepath.Join(currentStateDir, file.Name())
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Printf("[ns] Warning: failed to read container file %s: %v\n", filePath, err)
			continue
		}

		var containerInfo ContainerInfo
		if err := json.Unmarshal(data, &containerInfo); err != nil {
			fmt.Printf("[ns] Warning: failed to parse container file %s: %v\n", filePath, err)
			continue
		}

		// Check if the process is still running
		if isProcessRunning(containerInfo.PID) {
			containerInfo.Status = "running"
		} else {
			containerInfo.Status = "exited"
			// Clean up dead container info
			go func(id string) {
				UnregisterContainer(id)
			}(containerInfo.ID)
		}

		containers = append(containers, containerInfo)
	}

	return containers, nil
}

// isProcessRunning checks if a process with the given PID is still running
func isProcessRunning(pid int) bool {
	// Try to send signal 0 to the process (doesn't actually send a signal, just checks if process exists)
	err := syscall.Kill(pid, 0)
	return err == nil
}

// GetContainerByPID finds a container by its PID
func GetContainerByPID(pid int) (*ContainerInfo, error) {
	containers, err := ListContainers()
	if err != nil {
		return nil, err
	}

	for _, container := range containers {
		if container.PID == pid {
			return &container, nil
		}
	}

	return nil, fmt.Errorf("container with PID %d not found", pid)
}

// FormatContainerTable formats container information as a table
func FormatContainerTable(containers []ContainerInfo) string {
	if len(containers) == 0 {
		return "No containers found.\n"
	}

	// Header
	output := fmt.Sprintf("%-20s %-8s %-10s %-20s %-30s\n",
		"CONTAINER ID", "PID", "STATUS", "STARTED", "COMMAND")
	output += strings.Repeat("-", 90) + "\n"

	// Container rows
	for _, container := range containers {
		// Format start time
		startTime := container.StartTime.Format("15:04:05")

		// Build command string
		commandStr := container.Command
		if len(container.Args) > 0 {
			commandStr += " " + strings.Join(container.Args, " ")
		}

		// Truncate command if too long
		if len(commandStr) > 28 {
			commandStr = commandStr[:25] + "..."
		}

		// Truncate container ID for display
		displayID := container.ID
		if len(displayID) > 18 {
			displayID = displayID[:15] + "..."
		}

		output += fmt.Sprintf("%-20s %-8d %-10s %-20s %-30s\n",
			displayID, container.PID, container.Status, startTime, commandStr)
	}

	return output
}
