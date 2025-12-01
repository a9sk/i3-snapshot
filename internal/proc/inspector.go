package proc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetCommandFromPID returns the command line used to start the process with the given PID.
// It reads /proc/[PID]/cmdline and converts the null-separated content into a space-separated string.
func GetCommandFromPID(pid int) (string, error) {
	if pid <= 0 {
		return "", fmt.Errorf("invalid pid: %d", pid)
	}

	cmdlinePath := filepath.Join("/proc", fmt.Sprintf("%d", pid), "cmdline")
	data, err := os.ReadFile(cmdlinePath)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", cmdlinePath, err)
	}

	// /proc/[pid]/cmdline is null-byte separated; trim any trailing null and join with spaces
	parts := strings.Split(strings.TrimRight(string(data), "\x00"), "\x00")
	return strings.Join(parts, " "), nil
}

// GetCWDFromPID returns the current working directory of the given PID by resolving /proc/[PID]/cwd.
func GetCWDFromPID(pid int) (string, error) {
	if pid <= 0 {
		return "", fmt.Errorf("invalid pid: %d", pid)
	}

	cwdPath := filepath.Join("/proc", fmt.Sprintf("%d", pid), "cwd")
	dir, err := os.Readlink(cwdPath)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", cwdPath, err)
	}
	return dir, nil
}
