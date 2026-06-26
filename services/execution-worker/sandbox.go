package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// SandboxEnvironment represents the type of sandbox execution environment.
type SandboxEnvironment string

const (
	SandboxDocker     SandboxEnvironment = "docker"
	SandboxContainerd SandboxEnvironment = "containerd-runC"
	SandboxLocal      SandboxEnvironment = "local"
)

// SandboxConfig holds configuration for sandbox execution.
type SandboxConfig struct {
	// Environments is the ordered list of sandbox environments to try.
	Environments []SandboxEnvironment
	// DockerImage is the Docker image to use for sandbox execution.
	DockerImage string
	// ResourceLimits defines the resource constraints.
	ResourceLimits ResourceLimits
	// GracePeriod is additional time before SIGKILL after timeout.
	GracePeriod time.Duration
	// MaxOutputSize is the maximum allowed output size in bytes.
	MaxOutputSize int64
	// EnableLocalFallback enables local execution fallback when sandbox is unavailable.
	EnableLocalFallback bool
	// WorkingDir is the working directory for local execution.
	WorkingDir string
	// WarmPoolSize is the number of warm containers to keep ready.
	WarmPoolSize int
	// CompiledBinCacheSize is the max number of cached compiled binaries.
	CompiledBinCacheSize int
}

// DefaultSandboxConfig returns a default sandbox configuration.
func DefaultSandboxConfig() SandboxConfig {
	return SandboxConfig{
		Environments: []SandboxEnvironment{
			SandboxDocker,
			SandboxContainerd,
			SandboxLocal,
		},
		DockerImage:          getEnv("SANDBOX_IMAGE", "golang:1.21-alpine"),
		ResourceLimits:       DefaultResourceLimits(DifficultyMedium),
		GracePeriod:          2 * time.Second,
		MaxOutputSize:        1 * 1024 * 1024, // 1MB
		EnableLocalFallback:  getEnvBool("ENABLE_LOCAL_FALLBACK", false),
		WorkingDir:           getEnv("SANDBOX_WORKDIR", "/tmp/sandbox"),
		WarmPoolSize:         getEnvInt("WARM_POOL_SIZE", 2),
		CompiledBinCacheSize: 20,
	}
}

// SandboxResult holds the result of a sandboxed execution.
type SandboxResult struct {
	Stdout        string        `json:"stdout"`
	Stderr        string        `json:"stderr"`
	ExitCode      int           `json:"exit_code"`
	ExecutionTime time.Duration `json:"execution_time"`
	MemoryUsedKB  int           `json:"memory_used_kb"`
	CPUTimeMs     int           `json:"cpu_time_ms"`
	DiskIOKB      int           `json:"disk_io_kb"`
	NetworkIOKB   int           `json:"network_io_kb"`
	TimedOut      bool          `json:"timed_out"`
	ErrorMessage  string        `json:"error_message,omitempty"`
	SandboxEnv    string        `json:"sandbox_env"`
}

// SandboxMetrics holds resource usage metrics collected during execution.
type SandboxMetrics struct {
	MemoryPeakKB  int     `json:"memory_peak_kb"`
	CPUTimeMs     int     `json:"cpu_time_ms"`
	DiskReadKB    int64   `json:"disk_read_kb"`
	DiskWriteKB   int64   `json:"disk_write_kb"`
	NetworkRxKB   int64   `json:"network_rx_kb"`
	NetworkTxKB   int64   `json:"network_tx_kb"`
	OutputSize    int     `json:"output_size"`
}

// SandboxExecutor handles code execution in isolated environments.
type SandboxExecutor struct {
	config     SandboxConfig
	mu         sync.Mutex
	pool       *ContainerPool
	memory     *SandboxMemoryManager
}

// NewSandboxExecutor creates a new sandbox executor with the given configuration.
func NewSandboxExecutor(config SandboxConfig) *SandboxExecutor {
	return &SandboxExecutor{
		config: config,
		pool:   NewContainerPool(config.WarmPoolSize, config.DockerImage),
		memory: NewSandboxMemoryManager(config.CompiledBinCacheSize),
	}
}

// Execute runs the given code in a sandboxed environment with per-test-case timeout.
// It tries environments in order (Docker -> containerd -> local) for graceful degradation.
func (s *SandboxExecutor) Execute(ctx context.Context, code string, language string, timeout time.Duration) (*SandboxResult, error) {
	// Try each sandbox environment in order
	lastErr := fmt.Errorf("no sandbox environment available")

	for _, env := range s.config.Environments {
		// Skip Docker if disabled
		if env == SandboxDocker && !s.config.EnableLocalFallback {
			if !isDockerAvailable() {
				log.Printf("Docker unavailable, trying next environment")
				continue
			}
		}

		var result *SandboxResult
		var err error

		switch env {
		case SandboxDocker:
			result, err = s.executeInDocker(ctx, code, language, timeout)
		case SandboxContainerd:
			result, err = s.executeInContainerd(ctx, code, language, timeout)
		case SandboxLocal:
			if s.config.EnableLocalFallback {
				result, err = s.executeLocal(ctx, code, language, timeout)
			} else {
				continue
			}
		}

		if err == nil {
			result.SandboxEnv = string(env)
			return result, nil
		}
		lastErr = err
		log.Printf("Sandbox environment %s failed: %v", env, err)
	}

	return &SandboxResult{
		ErrorMessage: fmt.Sprintf("all sandbox environments failed: %v", lastErr),
	}, fmt.Errorf("sandbox execution failed: %v", lastErr)
}

// executeInDocker runs the code inside a Docker container with security hardening.
func (s *SandboxExecutor) executeInDocker(ctx context.Context, code string, language string, timeout time.Duration) (*SandboxResult, error) {
	if !isDockerAvailable() {
		return nil, fmt.Errorf("Docker is not available")
	}

	// Try to use a warm container from the pool
	containerID, err := s.pool.Acquire(ctx)
	if err != nil {
		// Fall back to creating a new container
		containerID = ""
	}

	// Create a temporary directory for the code
	tmpDir, err := os.MkdirTemp(s.config.WorkingDir, "exec-*")
	if err != nil {
		if containerID != "" {
			s.pool.Release(containerID, false)
		}
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Write code to file
	codeFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(codeFile, []byte(code), 0644); err != nil {
		os.RemoveAll(tmpDir)
		if containerID != "" {
			s.pool.Release(containerID, false)
		}
		return nil, fmt.Errorf("failed to write code file: %w", err)
	}

	// Write seccomp profile for Go binary
	seccompFile := filepath.Join(tmpDir, "seccomp-go.json")
	if err := writeSeccompProfile(seccompFile); err != nil {
		log.Printf("Warning: failed to write seccomp profile: %v", err)
	}

	containerName := fmt.Sprintf("exec-%d", time.Now().UnixNano())

	// Security-hardened Docker arguments
	args := []string{
		"run",
		"--rm",
		"--name", containerName,
		"--network", "none",                                          // Network egress deny
		"--dns", "8.8.8.8",                                           // DNS exception (not used with --network=none, but for reference)
		"--read-only",                                                // Read-only root filesystem
		"--tmpfs", "/tmp:size=50m,noexec,nosuid,nodev",              // Writable /tmp with restrictions
		"--tmpfs", "/go:size=100m,noexec,nosuid,nodev",              // Writable /go cache
		"--security-opt", "no-new-privileges:true",                   // Prevent privilege escalation
		"--security-opt", fmt.Sprintf("seccomp=%s", seccompFile),     // Seccomp profile
		"--cap-drop", "ALL",                                          // Drop all capabilities
		"--cap-add", "DAC_OVERRIDE",                                  // Minimal required for Go
		fmt.Sprintf("--memory=%dm", s.config.ResourceLimits.MemoryLimitMB),
		fmt.Sprintf("--memory-swap=%dm", s.config.ResourceLimits.MemoryLimitMB), // No swap
		fmt.Sprintf("--cpus=%s", s.config.ResourceLimits.CPUQuota),
		fmt.Sprintf("--pids-limit=%d", s.config.ResourceLimits.PIDsLimit),
		"--ulimit", "nofile=1024:1024",                               // Limit file descriptors
		"--ulimit", "nproc=100:100",                                  // Limit processes
		"--label", "sandbox-type=code-execution",
	}

	// Add AppArmor profile if available
	if isAppArmorAvailable() {
		args = append(args, "--security-opt", "apparmor=sandbox-default")
	}

	// Mount code directory
	args = append(args, "-v", fmt.Sprintf("%s:/app:ro", tmpDir)) // Read-only mount for code
	args = append(args, "-v", fmt.Sprintf("%s:/seccomp:ro", tmpDir))
	args = append(args, "-w", "/app")
	args = append(args, s.config.DockerImage)

	// Compile first, then run (for caching)
	args = append(args, "sh", "-c", "go build -o /tmp/main main.go && /tmp/main")

	// Create context with per-test-case timeout + grace period
	totalTimeout := timeout + s.config.GracePeriod
	execCtx, cancel := context.WithTimeout(ctx, totalTimeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "docker", args...)

	// Capture output with size limiting
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &sanitizedWriter{inner: &limitedWriter{w: &stdout, limit: s.config.MaxOutputSize}}
	cmd.Stderr = &sanitizedWriter{inner: &limitedWriter{w: &stderr, limit: s.config.MaxOutputSize}}

	// Start resource monitoring
	monitorCtx, monitorCancel := context.WithCancel(ctx)
	defer monitorCancel()
	metrics := s.monitorContainerResources(monitorCtx, containerName, timeout)

	// Execute
	startTime := time.Now()
	err = cmd.Run()
	executionTime := time.Since(startTime)
	monitorCancel()

	result := &SandboxResult{
		Stdout:        stdout.String(),
		Stderr:        stderr.String(),
		ExecutionTime: executionTime,
	}

	// Check for timeout
	if execCtx.Err() == context.DeadlineExceeded || executionTime > timeout+s.config.GracePeriod {
		result.TimedOut = true
		result.ErrorMessage = fmt.Sprintf("execution exceeded time limit of %v", timeout)
		// SIGKILL after grace period
		killCmd := exec.Command("docker", "kill", "--signal", "SIGKILL", containerName)
		killCmd.Run()
	}

	// Set exit code
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
		if result.ErrorMessage == "" {
			result.ErrorMessage = sanitizeOutput(stderr.String())
		}
	}

	// Collect metrics
	if metrics != nil {
		result.MemoryUsedKB = metrics.MemoryPeakKB
		result.CPUTimeMs = metrics.CPUTimeMs
		result.DiskIOKB = int(metrics.DiskReadKB + metrics.DiskWriteKB)
	}

	// Return container to pool if warm
	if containerID != "" {
		s.pool.Release(containerID, result.ExitCode == 0)
	}

	// Cleanup temp files
	os.RemoveAll(tmpDir)

	return result, nil
}

// executeInContainerd runs the code using containerd/runC with security hardening.
func (s *SandboxExecutor) executeInContainerd(ctx context.Context, code string, language string, timeout time.Duration) (*SandboxResult, error) {
	// Check if containerd is available
	if !isContainerdAvailable() {
		return nil, fmt.Errorf("containerd is not available")
	}

	tmpDir, err := os.MkdirTemp(s.config.WorkingDir, "ctr-exec-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	codeFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(codeFile, []byte(code), 0644); err != nil {
		return nil, fmt.Errorf("failed to write code file: %w", err)
	}

	// Build binary first
	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(tmpDir, "main"), codeFile)
	if buildOutput, err := buildCmd.CombinedOutput(); err != nil {
		return &SandboxResult{
			ExitCode:     1,
			ErrorMessage: fmt.Sprintf("compilation error: %s", sanitizeOutput(string(buildOutput))),
		}, nil
	}

	// Use ctr (containerd CLI) to run with runC
	args := []string{
		"run",
		"--rm",
		"--runtime", "io.containerd.runc.v2",
		"--read-only",
		"--tmpfs", "/tmp:size=50m",
		fmt.Sprintf("--memory-limit=%d", s.config.ResourceLimits.MemoryLimitMB*1024*1024),
		fmt.Sprintf("--cpus=%s", s.config.ResourceLimits.CPUQuota),
		"--env", "GOCACHE=/tmp/gocache",
		"--mount", fmt.Sprintf("type=bind,src=%s,dst=/app,options=ro", tmpDir),
		s.config.DockerImage,
		fmt.Sprintf("exec-container-%d", time.Now().UnixNano()),
		"/app/main",
	}

	totalTimeout := timeout + s.config.GracePeriod
	execCtx, cancel := context.WithTimeout(ctx, totalTimeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "ctr", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &sanitizedWriter{inner: &limitedWriter{w: &stdout, limit: s.config.MaxOutputSize}}
	cmd.Stderr = &sanitizedWriter{inner: &limitedWriter{w: &stderr, limit: s.config.MaxOutputSize}}

	startTime := time.Now()
	err = cmd.Run()
	executionTime := time.Since(startTime)

	result := &SandboxResult{
		Stdout:        stdout.String(),
		Stderr:        stderr.String(),
		ExecutionTime: executionTime,
		SandboxEnv:    string(SandboxContainerd),
	}

	if execCtx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		result.ErrorMessage = fmt.Sprintf("execution exceeded time limit of %v", timeout)
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
		if result.ErrorMessage == "" {
			result.ErrorMessage = sanitizeOutput(stderr.String())
		}
	}

	return result, nil
}

// executeLocal runs the code locally with resource monitoring and output sanitization.
func (s *SandboxExecutor) executeLocal(ctx context.Context, code string, language string, timeout time.Duration) (*SandboxResult, error) {
	tmpDir, err := os.MkdirTemp(s.config.WorkingDir, "local-exec-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	codeFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(codeFile, []byte(code), 0644); err != nil {
		return nil, fmt.Errorf("failed to write code file: %w", err)
	}

	totalTimeout := timeout + s.config.GracePeriod
	execCtx, cancel := context.WithTimeout(ctx, totalTimeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "go", "run", codeFile)
	cmd.Dir = tmpDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &sanitizedWriter{inner: &limitedWriter{w: &stdout, limit: s.config.MaxOutputSize}}
	cmd.Stderr = &sanitizedWriter{inner: &limitedWriter{w: &stderr, limit: s.config.MaxOutputSize}}

	// Run with resource monitoring
	// (SysProcAttr resource limits are platform-specific; Docker is used for real isolation)

	startTime := time.Now()
	err = cmd.Run()
	executionTime := time.Since(startTime)

	result := &SandboxResult{
		Stdout:        stdout.String(),
		Stderr:        stderr.String(),
		ExecutionTime: executionTime,
		SandboxEnv:    string(SandboxLocal),
	}

	if execCtx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		result.ErrorMessage = fmt.Sprintf("execution exceeded time limit of %v", timeout)
		// Force kill
		if cmd.Process != nil {
			cmd.Process.Signal(os.Kill)
		}
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
		if result.ErrorMessage == "" {
			result.ErrorMessage = sanitizeOutput(stderr.String())
		}
	}

	// Estimate memory usage from Go runtime
	result.MemoryUsedKB = estimateMemoryUsage(cmd)

	return result, nil
}

// monitorContainerResources monitors Docker container resource usage during execution.
func (s *SandboxExecutor) monitorContainerResources(ctx context.Context, containerName string, timeout time.Duration) *SandboxMetrics {
	metrics := &SandboxMetrics{}
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return metrics
		case <-ticker.C:
			// Get Docker stats
			stats, err := getDockerContainerStats(ctx, containerName)
			if err != nil {
				continue
			}
			if stats.MemoryPeakKB > metrics.MemoryPeakKB {
				metrics.MemoryPeakKB = stats.MemoryPeakKB
			}
			if stats.CPUTimeMs > metrics.CPUTimeMs {
				metrics.CPUTimeMs = stats.CPUTimeMs
			}
			metrics.DiskReadKB += stats.DiskReadKB
			metrics.DiskWriteKB += stats.DiskWriteKB
			metrics.NetworkRxKB += stats.NetworkRxKB
			metrics.NetworkTxKB += stats.NetworkTxKB
		}
	}
}

// CompileAndCache compiles the code and caches the binary for reuse.
func (s *SandboxExecutor) CompileAndCache(ctx context.Context, code string) (string, error) {
	return s.memory.Compile(ctx, code, s.config.DockerImage, s.config.WorkingDir)
}

// getDockerContainerStats retrieves real-time resource usage for a Docker container.
func getDockerContainerStats(ctx context.Context, containerID string) (*SandboxMetrics, error) {
	cmd := exec.CommandContext(ctx, "docker", "stats", "--no-stream", "--format",
		"{{.MemUsage}}|{{.CPUPerc}}|{{.NetIO}}|{{.BlockIO}}", containerID)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseDockerStats(string(output))
}

// parseDockerStats parses Docker stats output into metrics.
func parseDockerStats(stats string) (*SandboxMetrics, error) {
	parts := strings.Split(strings.TrimSpace(stats), "|")
	if len(parts) < 4 {
		return nil, fmt.Errorf("unexpected docker stats format: %s", stats)
	}

	metrics := &SandboxMetrics{}

	// Parse memory (e.g., "10MiB / 100MiB")
	memParts := strings.Split(parts[0], "/")
	memUsed := strings.TrimSpace(memParts[0])
	metrics.MemoryPeakKB = parseMemoryToKB(memUsed)

	// Parse CPU percentage (e.g., "12.5%")
	cpuStr := strings.TrimSuffix(strings.TrimSpace(parts[1]), "%")
	if cpuVal, err := strconv.ParseFloat(cpuStr, 64); err == nil {
		metrics.CPUTimeMs = int(cpuVal * 10) // Approximate
	}

	// Parse network I/O (e.g., "10kB / 5kB")
	netParts := strings.Split(parts[2], "/")
	if len(netParts) >= 2 {
		metrics.NetworkRxKB = parseTrafficToKB(strings.TrimSpace(netParts[0]))
		metrics.NetworkTxKB = parseTrafficToKB(strings.TrimSpace(netParts[1]))
	}

	// Parse block I/O (e.g., "100kB / 50kB")
	ioParts := strings.Split(parts[3], "/")
	if len(ioParts) >= 2 {
		metrics.DiskReadKB = parseTrafficToKB(strings.TrimSpace(ioParts[0]))
		metrics.DiskWriteKB = parseTrafficToKB(strings.TrimSpace(ioParts[1]))
	}

	return metrics, nil
}

// parseMemoryToKB converts a memory string (e.g., "10MiB") to KB.
func parseMemoryToKB(memStr string) int {
	memStr = strings.TrimSpace(memStr)
	if memStr == "" || memStr == "--" {
		return 0
	}

	var value float64
	var unit string
	fmt.Sscanf(memStr, "%f%s", &value, &unit)

	switch strings.ToUpper(strings.TrimSpace(unit)) {
	case "KIB":
		return int(value)
	case "MIB":
		return int(value * 1024)
	case "GIB":
		return int(value * 1024 * 1024)
	case "KB":
		return int(value)
	case "MB":
		return int(value * 1024)
	case "GB":
		return int(value * 1024 * 1024)
	case "B":
		return int(value / 1024)
	default:
		return int(value) // Assume KB
	}
}

// parseTrafficToKB converts a traffic string (e.g., "10kB") to KB.
func parseTrafficToKB(trafficStr string) int64 {
	trafficStr = strings.TrimSpace(trafficStr)
	if trafficStr == "" || trafficStr == "--" || trafficStr == "0B" {
		return 0
	}

	var value float64
	var unit string
	fmt.Sscanf(trafficStr, "%f%s", &value, &unit)

	switch strings.ToUpper(strings.TrimSpace(unit)) {
	case "KB", "KIB":
		return int64(value)
	case "MB", "MIB":
		return int64(value * 1024)
	case "GB", "GIB":
		return int64(value * 1024 * 1024)
	case "B":
		return int64(value / 1024)
	default:
		return int64(value)
	}
}

// writeSeccompProfile writes a seccomp profile for Go binary execution to a file.
func writeSeccompProfile(path string) error {
	profile := `{
	"defaultAction": "SCMP_ACT_ERRNO",
	"architectures": ["SCMP_ARCH_X86_64", "SCMP_ARCH_AARCH64"],
	"syscalls": [
		{
			"names": [
				"accept4", "access", "arch_prctl", "bind", "brk", "capget", "capset",
				"chdir", "chmod", "clock_getres", "clock_gettime", "clock_nanosleep",
				"clone", "close", "connect", "copy_file_range", "creat",
				"dup", "dup2", "dup3", "epoll_create1", "epoll_ctl", "epoll_pwait",
				"eventfd2", "execve", "exit", "exit_group", "faccessat2",
				"fchdir", "fchmod", "fchmodat2", "fchown", "fchownat",
				"fcntl", "fdatasync", "fgetxattr", "flistxattr",
				"flock", "fremovexattr", "fsetxattr", "fstat", "fstatfs",
				"fsync", "ftruncate", "futex", "futimens", "getcpu",
				"getcwd", "getdents64", "getegid", "geteuid", "getgid",
				"getgroups", "getpeername", "getpgid", "getpid", "getppid",
				"getpriority", "getrandom", "getresgid", "getresuid",
				"getrlimit", "getrusage", "getsockname", "getsockopt",
				"gettid", "gettimeofday", "getuid", "getxattr",
				"inotify_add_watch", "inotify_init1", "inotify_rm_watch",
				"ioctl", "ioprio_get", "ioprio_set", "ipc", "listen",
				"lgetxattr", "link", "linkat", "listxattr", "llistxattr",
				"lremovexattr", "lseek", "lsetxattr", "lstat", "madvise",
				"mbind", "memfd_create", "membarrier", "mincore",
				"mkdir", "mkdirat", "mlock", "mlock2", "mlockall",
				"mmap", "mount", "mprotect", "mremap", "msgctl", "msgget",
				"msgrcv", "msgsnd", "msync", "munlock", "munlockall",
				"munmap", "name_to_handle_at", "nanosleep", "newfstatat",
				"open", "openat", "openat2", "pause", "pidfd_open",
				"pipe", "pipe2", "pkey_alloc", "pkey_free", "pkey_mprotect",
				"poll", "ppoll", "prctl", "pread64", "preadv", "preadv2",
				"prlimit64", "process_vm_readv", "process_vm_writev",
				"pselect6", "pwrite64", "pwritev", "pwritev2",
				"read", "readlink", "readlinkat", "readv", "reboot",
				"recvfrom", "recvmmsg", "recvmsg", "rename", "renameat",
				"renameat2", "rmdir", "rseq", "rt_sigaction",
				"rt_sigpending", "rt_sigprocmask", "rt_sigqueueinfo",
				"rt_sigreturn", "rt_sigsuspend", "rt_sigtimedwait",
				"rt_tgsigqueueinfo", "sched_getaffinity", "sched_getattr",
				"sched_getparam", "sched_getscheduler", "sched_rr_get_interval",
				"sched_setaffinity", "sched_setattr", "sched_setparam",
				"sched_setscheduler", "sched_yield", "seccomp",
				"select", "semctl", "semget", "semop", "semtimedop",
				"sendfile", "sendmmsg", "sendmsg", "sendto", "set_robust_list",
				"set_tid_address", "setgid", "setgroups", "setitimer",
				"setpgid", "setpriority", "setregid", "setresgid",
				"setresuid", "setreuid", "setrlimit", "setsid",
				"setsockopt", "setuid", "setxattr", "shmat", "shmctl",
				"shmdt", "shmget", "shutdown", "sigaltstack", "signalfd4",
				"socket", "socketpair", "splice", "stat", "statfs",
				"statx", "symlink", "symlinkat", "sync", "sync_file_range",
				"syncfs", "sysinfo", "tee", "tgkill", "time",
				"timer_create", "timer_delete", "timer_getoverrun",
				"timer_gettime", "timer_settime", "timerfd_create",
				"timerfd_gettime", "timerfd_settime", "times",
				"truncate", "ugetrlimit", "umask", "uname", "unlink",
				"unlinkat", "unshare", "utimensat", "utimes",
				"wait4", "waitid", "write", "writev"
			],
			"action": "SCMP_ACT_ALLOW"
		}
	]
}`

	return os.WriteFile(path, []byte(profile), 0644)
}

// isDockerAvailable checks if Docker is available on the system.
func isDockerAvailable() bool {
	cmd := exec.Command("docker", "info")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run() == nil
}

// isContainerdAvailable checks if containerd CLI is available.
func isContainerdAvailable() bool {
	cmd := exec.Command("ctr", "--version")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run() == nil
}

// isAppArmorAvailable checks if AppArmor is available.
func isAppArmorAvailable() bool {
	cmd := exec.Command("which", "aa-status")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run() == nil
}

// getSysProcAttr returns system process attributes for resource limiting.
func getSysProcAttr(limits ResourceLimits) interface{} {
	return nil // Platform-specific; use syscall.SysProcAttr on Linux
}

// estimateMemoryUsage estimates memory usage of a completed process.
func estimateMemoryUsage(cmd *exec.Cmd) int {
	if cmd.ProcessState == nil {
		return 0
	}
	rusage := cmd.ProcessState.SysUsage()
	if rusage == nil {
		return 0
	}
	if v, ok := rusage.(interface{ MaxRSS() int64 }); ok {
		return int(v.MaxRSS())
	}
	return 0
}

// ============================================================================
// Output Sanitization
// ============================================================================

// sanitizeOutput strips ANSI escape codes and control characters from output.
func sanitizeOutput(output string) string {
	// Strip ANSI escape sequences
	ansiRe := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	output = ansiRe.ReplaceAllString(output, "")

	// Strip other escape sequences (OSC, etc.)
	oscRe := regexp.MustCompile(`\x1b\].*?(\x1b\\|\x07)`)
	output = oscRe.ReplaceAllString(output, "")

	// Strip control characters except newline, tab, carriage return
	ctrlRe := regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
	output = ctrlRe.ReplaceAllString(output, "")

	return output
}

// sanitizedWriter wraps a writer, sanitizing all output before writing.
type sanitizedWriter struct {
	inner io.Writer
}

func (w *sanitizedWriter) Write(p []byte) (int, error) {
	sanitized := sanitizeOutput(string(p))
	return w.inner.Write([]byte(sanitized))
}

// limitedWriter limits the amount of data written.
type limitedWriter struct {
	w     io.Writer
	written int64
	limit   int64
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	if w.written >= w.limit {
		return 0, fmt.Errorf("output size limit exceeded")
	}
	remaining := w.limit - w.written
	if int64(len(p)) > remaining {
		p = p[:remaining]
	}
	n, err := w.w.Write(p)
	w.written += int64(n)
	return n, err
}

// ============================================================================
// Sandbox Escape Detection
// ============================================================================

// detectSandboxEscape checks if the output contains signs of sandbox escape attempts.
func detectSandboxEscape(output string) (bool, string) {
	lower := strings.ToLower(output)

	// Suspicious patterns
	escapePatterns := []struct {
		pattern string
		message string
	}{
		{"/proc/1/cmdline", "attempt to read host process info"},
		{"/etc/shadow", "attempt to read sensitive host file"},
		{"/etc/passwd", "attempt to read host password file"},
		{"/var/run/docker.sock", "attempt to access Docker socket"},
		{"docker", "potential Docker escape attempt"},
		{"--privileged", "privilege escalation attempt"},
		{"cap_sys_admin", "capability escalation attempt"},
		{"modprobe", "kernel module loading attempt"},
		{"nsenter", "namespace escape attempt"},
		{"chroot", "chroot escape attempt"},
		{"mount --bind", "bind mount escape attempt"},
		{"cgroup", "cgroup escape attempt"},
		{"ptrace", "process tracing escape attempt"},
	}

	for _, ep := range escapePatterns {
		if strings.Contains(lower, ep.pattern) {
			return true, ep.message
		}
	}

	return false, ""
}

// ============================================================================
// Container Pool for warm containers
// ============================================================================

// ContainerPool manages a pool of warm containers for faster startup.
type ContainerPool struct {
	mu        sync.Mutex
	size      int
	image     string
	available []string
	running   int
}

// NewContainerPool creates a new container pool.
func NewContainerPool(size int, image string) *ContainerPool {
	return &ContainerPool{
		size:      size,
		image:     image,
		available: make([]string, 0, size),
	}
}

// Acquire gets a warm container from the pool, or returns empty string if none available.
func (p *ContainerPool) Acquire(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.available) > 0 {
		id := p.available[len(p.available)-1]
		p.available = p.available[:len(p.available)-1]
		return id, nil
	}

	return "", fmt.Errorf("no warm containers available")
}

// Release returns a container to the pool or stops it.
func (p *ContainerPool) Release(containerID string, reusable bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if reusable && len(p.available) < p.size {
		p.available = append(p.available, containerID)
	} else {
		// Stop the container
		exec.Command("docker", "rm", "-f", containerID).Run()
	}
}

// ============================================================================
// SandboxMemoryManager for compiled binary caching
// ============================================================================

// SandboxMemoryManager caches compiled binaries to avoid recompilation.
type SandboxMemoryManager struct {
	mu    sync.Mutex
	cache map[string]*CachedBinary
	size  int
}

// CachedBinary holds a compiled binary and its metadata.
type CachedBinary struct {
	Path      string
	CreatedAt time.Time
	Hash      string
	Hits      int
}

// NewSandboxMemoryManager creates a new binary cache.
func NewSandboxMemoryManager(size int) *SandboxMemoryManager {
	return &SandboxMemoryManager{
		cache: make(map[string]*CachedBinary),
		size:  size,
	}
}

// Compile compiles the code and caches the binary, or returns cached path.
func (m *SandboxMemoryManager) Compile(ctx context.Context, code, image, workDir string) (string, error) {
	codeHash := fmt.Sprintf("%x", []byte(code))
	if len(codeHash) > 32 {
		codeHash = codeHash[:32]
	}

	m.mu.Lock()
	if cached, ok := m.cache[codeHash]; ok {
		cached.Hits++
		m.mu.Unlock()
		return cached.Path, nil
	}
	m.mu.Unlock()

	// Compile
	tmpDir, err := os.MkdirTemp(workDir, "compile-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	codeFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(codeFile, []byte(code), 0644); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to write code: %w", err)
	}

	binaryPath := filepath.Join(tmpDir, "main")
	cmd := exec.CommandContext(ctx, "go", "build", "-o", binaryPath, codeFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("compilation failed: %s", sanitizeOutput(string(output)))
	}

	cached := &CachedBinary{
		Path:      binaryPath,
		CreatedAt: time.Now(),
		Hash:      codeHash,
	}

	m.mu.Lock()
	// Evict oldest if full
	if len(m.cache) >= m.size {
		var oldestKey string
		var oldestTime time.Time
		for k, v := range m.cache {
			if oldestKey == "" || v.CreatedAt.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.CreatedAt
			}
		}
		if oldestKey != "" {
			os.RemoveAll(filepath.Dir(m.cache[oldestKey].Path))
			delete(m.cache, oldestKey)
		}
	}
	m.cache[codeHash] = cached
	m.mu.Unlock()

	return binaryPath, nil
}

// Stats returns cache statistics.
func (m *SandboxMemoryManager) Stats() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	totalHits := 0
	for _, c := range m.cache {
		totalHits += c.Hits
	}

	return map[string]interface{}{
		"cache_size": len(m.cache),
		"max_size":   m.size,
		"total_hits": totalHits,
	}
}
