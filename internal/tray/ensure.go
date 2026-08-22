package tray

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"zenpomo/internal/daemon"
)

// IsTrayRunning checks if a ZenPomo system tray process is already active.
func IsTrayRunning() bool {
	currentPid := os.Getpid()

	if runtime.GOOS == "linux" {
		entries, err := os.ReadDir("/proc")
		if err != nil {
			return false
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			pid, err := strconv.Atoi(entry.Name())
			if err != nil || pid == currentPid {
				continue
			}

			cmdlineBytes, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "cmdline"))
			if err != nil {
				continue
			}

			cmdline := string(bytes.ReplaceAll(cmdlineBytes, []byte{0}, []byte(" ")))
			if strings.Contains(cmdline, "zenpomo tray") && !strings.Contains(cmdline, "grep") {
				return true
			}
		}
	} else if runtime.GOOS == "windows" {
		cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq zenpomo.exe")
		out, err := cmd.Output()
		if err == nil && strings.Contains(string(out), "zenpomo.exe") {
			return true
		}
	}

	return false
}

// EnsureTray checks if the system tray process is running, and starts it in background if not.
func EnsureTray() {
	if IsTrayRunning() {
		return
	}

	exe, err := os.Executable()
	if err != nil {
		exe = "zenpomo"
	}

	cmd := exec.Command(exe, "tray")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	daemon.SetDetachedProcess(cmd)

	_ = cmd.Start()
}
