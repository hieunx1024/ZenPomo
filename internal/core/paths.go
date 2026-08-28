package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// GetExecutable resolves the installed or current ZenPomo binary path, safely avoiding temporary test binaries.
func GetExecutable() string {
	// 1. Check user local installation directory
	home, err := os.UserHomeDir()
	if err == nil {
		var defaultBin string
		if runtime.GOOS == "windows" {
			localAppData := os.Getenv("LOCALAPPDATA")
			if localAppData != "" {
				defaultBin = filepath.Join(localAppData, "ZenPomo", "zenpomo.exe")
			}
		} else {
			defaultBin = filepath.Join(home, ".local", "bin", "zenpomo")
		}
		if defaultBin != "" {
			if _, err := os.Stat(defaultBin); err == nil {
				return defaultBin
			}
		}
	}

	// 2. Check running executable, ensuring it is not a temporary test binary
	exe, err := os.Executable()
	if err == nil && exe != "" {
		base := filepath.Base(exe)
		if !strings.HasSuffix(base, ".test") && !strings.Contains(exe, "go-build") && !strings.Contains(exe, "T___") {
			return exe
		}
	}

	// 3. Fallback to PATH lookup
	binName := "zenpomo"
	if runtime.GOOS == "windows" {
		binName = "zenpomo.exe"
	}
	if p, err := exec.LookPath(binName); err == nil {
		return p
	}

	return binName
}
