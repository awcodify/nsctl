# 🧭 Project Instructions — Minimal Container Runtime

### Project Overview

We’re building a **minimal container runtime** implemented in **Go**, using only **Linux namespaces** and **cgroups** — no Docker, no containerd, no complex dependencies.  
The goal is to understand and reimplement the essence of containerization in the simplest, most educational way.

---

## 🧱 Core Idea

A simple CLI tool called `nsctl` that can:

- Run commands in isolated Linux **namespaces**  
- Limit CPU and memory usage via **cgroups**  
- Mount `/proc` and set a custom hostname  
- Demonstrate how containers actually work internally  

Think of it as a lightweight, educational version of Docker.

---

## ⚙️ Tech Stack

- **Language:** Go  
- **Linux Concepts:** Namespaces, Cgroups  
- **Syscalls Used:**  
  - `Clone`, `Mount`, `Sethostname`, `Chroot`, `Execve`  
  - `os/exec` and `syscall.SysProcAttr` for namespace setup  
  - Manual writes to `/sys/fs/cgroup/...` for resource limits  

---

## 📂 Project Structure

```
nsctl/
├── cmd/
│   └── main.go         # CLI entrypoint
├── pkg/
│   ├── ns/             # Namespace logic
│   │   └── namespace.go
│   └── cgroup/         # Cgroup logic
│       └── cgroup.go
└── instructions.md     # Copilot instructions (this file)
```

---

## 🧩 Implementation Style

Copilot should:

- Write **clear, low-level, educational code**
- Prioritize **readability over abstraction**
- Use **standard library only**
- Always **print debug logs** (e.g. `"[ns] creating PID namespace"`)
- Add **explanatory comments** for every key action
- Avoid **Docker APIs**, **OCI libraries**, or **external frameworks**

---

## 🧠 Development Philosophy

1. **Build step-by-step** — one syscall or concept at a time  
2. **Explain with comments** — why each syscall or flag is used  
3. **Prefer explicit code** — no magic or abstraction layers  
4. **Log what’s happening** — for visibility and learning  
5. **Keep code minimal and educational**

---

## 🧰 Implementation Phases

### 1. Namespace Isolation
- Use `Clone()` or `SysProcAttr.Cloneflags`
- Create new **UTS**, **PID**, and **mount** namespaces
- Mount `/proc` inside the container
- Set hostname (e.g. `"container"`)

### 2. Cgroups (v2 or v1)
- Create a directory under `/sys/fs/cgroup/nsctl/<id>/`
- Write resource limits manually:
  ```bash
  echo "100000" > /sys/fs/cgroup/nsctl/demo/cpu.max
  echo "52428800" > /sys/fs/cgroup/nsctl/demo/memory.max
  ```
- Add container PID to `cgroup.procs`

### 3. CLI Interface
Example commands:
```bash
nsctl run -m 100m -cpu 0.5 /bin/bash
nsctl ps
```

- `run`: Create isolated process
- `ps`: Show running containers

### 4. (Optional Future)
- Overlay filesystem for isolation
- Networking namespace setup
- YAML config for resource presets

---

## 💬 Copilot Behavior Guidelines

Copilot should:
- Generate **Go code that reads like teaching material**
- Comment **why**, not just **what**
- Use **Go idioms**, not pseudo-C style
- Avoid clever abstractions, prefer clarity
- Use verbose variable names (e.g. `containerPID`, `cgroupPath`)
- Print out every system-level action (mounts, namespace creation, etc.)

Example of desired code/comment style:
```go
// Mount /proc inside the new PID namespace so commands like ps work correctly
err := syscall.Mount("proc", "/proc", "proc", 0, "")
if err != nil {
    log.Fatalf("Failed to mount /proc: %v", err)
}
```

---

## 🎨 Copilot Personality

When generating code, Copilot should act like a **pragmatic system engineer**:
- Think out loud through comments.
- Always justify syscalls.
- Avoid any mention of Docker or Kubernetes.
- Focus on how Linux primitives can build containers from scratch.
- Help explore internals, not build production systems.

---

## ✅ Example Output (After Phase 1)

```bash
$ go run cmd/main.go run /bin/bash
[ns] creating PID and UTS namespaces
[ns] mounting /proc
[ns] setting hostname to container
root@container:/#
```

At this stage, the project should produce an isolated shell with its own PID 1 process and `/proc` mount.

---

## 🧾 Summary

**Goal:** Build a functional, minimal container runtime to learn namespaces & cgroups  
**Language:** Go  
**Philosophy:** Pragmatic, minimal, educational  
**Style:** Explicit, commented, syscall-level clarity  

---
