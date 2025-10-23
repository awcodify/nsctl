# nsctl - Minimal Container Runtime

A simple educational container runtime implemented in Go using Linux namespaces and cgroups.

## Overview

`nsctl` demonstrates how containers work at the Linux kernel level by using:
- **UTS namespace** - Isolates hostname
- **PID namespace** - Isolates process IDs  
- **Mount namespace** - Isolates filesystem mounts

## Implementation

The core `run()` function in `pkg/ns/namespace.go` creates a new process with isolated namespaces:

```go
// RunSimple starts /bin/bash in isolated UTS, PID, and mount namespaces
func RunSimple() error {
    fmt.Printf("[ns] creating PID, UTS, and mount namespaces\n")
    
    // Create a new bash process
    cmd := exec.Command("/bin/bash")
    
    // Set up namespace isolation
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: unix.CLONE_NEWUTS | unix.CLONE_NEWPID | unix.CLONE_NEWNS,
    }
    
    // ... rest of implementation
}
```

## Usage

**Note: This only works on Linux** - namespaces are a Linux kernel feature.

### Build
```bash
go build -o nsctl ./cmd
```

### Run Commands

```bash
# Start a simple isolated bash shell
./nsctl simple

# Run a specific command in isolation  
./nsctl run /bin/bash

# List running containers (not implemented yet)
./nsctl ps
```

### Expected Output (Linux)
```bash
$ ./nsctl simple
[ns] creating PID, UTS, and mount namespaces
[ns] started bash with PID 1234 in isolated namespaces
root@container:/# 
```

Inside the container:
- `hostname` shows "container" 
- `ps` shows only processes in the isolated PID namespace
- Process runs as PID 1 in its namespace

## Architecture

```
nsctl/
├── cmd/main.go              # CLI entrypoint
├── pkg/ns/
│   ├── namespace.go         # Linux implementation (build constraint: linux)
│   └── namespace_stub.go    # Non-Linux stub (build constraint: !linux)
├── pkg/cgroup/              # Future: cgroup resource limits
└── go.mod
```

## Implementation Details

### Namespace Setup
- Uses `syscall.SysProcAttr.Cloneflags` with `exec.Command`
- Creates new UTS, PID, and mount namespaces via clone flags
- Connects stdin/stdout/stderr to parent process

### Future Enhancements
1. **Cgroups**: CPU/memory limits via `/sys/fs/cgroup/`
2. **Process Management**: Track running containers
3. **Filesystem Isolation**: chroot or overlay filesystems
4. **Network Namespaces**: Isolated networking

## Educational Goals

This project helps understand:
- How containers are just processes with Linux namespaces
- The syscalls underlying container runtimes like Docker
- Direct interaction with Linux kernel features
- Building system-level Go applications

## Limitations

- **Linux only** - uses Linux-specific syscalls
- **No filesystem isolation** - shares host filesystem  
- **No resource limits** - no cgroup integration yet
- **No networking** - uses host network
- **Educational purpose** - not production ready

## References

- [Linux Namespaces](https://man7.org/linux/man-pages/man7/namespaces.7.html)
- [Linux Cgroups](https://man7.org/linux/man-pages/man7/cgroups.7.html)
- [Container Internals](https://jvns.ca/blog/2016/10/10/what-even-is-a-container/)