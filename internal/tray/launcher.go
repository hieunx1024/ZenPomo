package tray

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"zenpomo/internal/daemon"
)

var (
	tuiProcessMu sync.Mutex
	activeTuiCmd *exec.Cmd
)

// FocusOrLaunchTUI brings the existing TUI window to the front or launches a new single instance.
func FocusOrLaunchTUI() error {
	tuiProcessMu.Lock()
	defer tuiProcessMu.Unlock()

	// If a TUI window is already running, focus it and NEVER spawn a duplicate
	if _, running := findRunningTUIProcess(); running {
		_ = focusExistingWindow()
		return nil
	}

	return launchNewTUI("tui")
}

// FocusOrLaunchConfig opens the Config modal in the existing TUI or launches a new single TUI instance.
func FocusOrLaunchConfig() error {
	tuiProcessMu.Lock()
	defer tuiProcessMu.Unlock()

	// 1. Send signal to daemon so any running TUI switches to Config mode in place
	client := daemon.NewClient()
	_ = client.EnsureDaemon()
	_, _ = client.SendCommand(daemon.CmdRequestConfig)

	// 2. If a TUI window is already running, focus it and do not spawn a duplicate
	if _, running := findRunningTUIProcess(); running {
		_ = focusExistingWindow()
		return nil
	}

	// 3. If no TUI is open, spawn a new window directly in config mode
	return launchNewTUI("config")
}

// LaunchOrToggleTUI toggles (opens if closed, closes if open) the floating TUI window.
func LaunchOrToggleTUI() error {
	tuiProcessMu.Lock()
	defer tuiProcessMu.Unlock()

	// If currently active and still running, close it (toggle behavior)
	if activeTuiCmd != nil && activeTuiCmd.Process != nil {
		_ = activeTuiCmd.Process.Signal(syscall.SIGTERM)
		activeTuiCmd = nil
		return nil
	}

	// Check if any external TUI process is running
	if pid, running := findRunningTUIProcess(); running {
		if proc, err := os.FindProcess(pid); err == nil {
			_ = proc.Signal(syscall.SIGTERM)
			return nil
		}
	}

	return launchNewTUI("tui")
}

func focusExistingWindow() bool {
	if runtime.GOOS == "windows" {
		return focusWindows()
	}
	return focusLinux()
}

func focusLinux() bool {
	// Try wmctrl first (X11 & XWayland)
	if p, err := exec.LookPath("wmctrl"); err == nil {
		if err := exec.Command(p, "-x", "-a", "ZenPomo").Run(); err == nil {
			return true
		}
		if err := exec.Command(p, "-a", "ZenPomo").Run(); err == nil {
			return true
		}
	}

	// Try xdotool
	if p, err := exec.LookPath("xdotool"); err == nil {
		out, err := exec.Command(p, "search", "--name", "ZenPomo").Output()
		if err == nil && len(bytes.TrimSpace(out)) > 0 {
			lines := strings.Split(strings.TrimSpace(string(out)), "\n")
			if len(lines) > 0 {
				winID := lines[len(lines)-1]
				if exec.Command(p, "windowactivate", "--sync", winID).Run() == nil {
					return true
				}
			}
		}
	}

	return false
}

func focusWindows() bool {
	if _, err := exec.LookPath("powershell.exe"); err == nil {
		script := `$w = Get-Process | Where-Object { $_.MainWindowTitle -like "*ZenPomo*" } | Select-Object -First 1; if ($w) { (New-Object -ComObject WScript.Shell).AppActivate($w.Id) }`
		return exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script).Run() == nil
	}
	return false
}

func findRunningTUIProcess() (int, bool) {
	currentPid := os.Getpid()

	if runtime.GOOS == "linux" {
		entries, err := os.ReadDir("/proc")
		if err != nil {
			return 0, false
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
			if (strings.Contains(cmdline, "zenpomo tui") || strings.Contains(cmdline, "zenpomo config") || (strings.Contains(cmdline, "zenpomo") && !strings.Contains(cmdline, "daemon") && !strings.Contains(cmdline, "tray") && !strings.Contains(cmdline, "status") && !strings.Contains(cmdline, "install") && !strings.Contains(cmdline, "autostart"))) && !strings.Contains(cmdline, "grep") {
				return pid, true
			}
		}
	}

	return 0, false
}

func launchNewTUI(appArgs ...string) error {
	exe, err := os.Executable()
	if err != nil {
		exe = "zenpomo"
	}

	if len(appArgs) == 0 {
		appArgs = []string{"tui"}
	}

	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		winArgs := append([]string{"--title", "ZenPomo", "-w", "0", "new-tab", "--size", "76,22", exe}, appArgs...)
		if _, err := exec.LookPath("wt.exe"); err == nil {
			cmd = exec.Command("wt.exe", winArgs...)
		} else {
			cmd = exec.Command("cmd.exe", "/c", "start", "ZenPomo", "mode", "con", "cols=76", "lines=22", "&&", exe, strings.Join(appArgs, " "))
		}
	} else {
		// Prepare terminal command arguments
		gnomeArgs := append([]string{"--class=ZenPomo", "--name=ZenPomo", "--title=ZenPomo", "--geometry=76x22", "--", exe}, appArgs...)
		alacrittyArgs := append([]string{"--class", "ZenPomo,ZenPomo", "--title", "ZenPomo", "-o", "window.dimensions.columns=76", "-o", "window.dimensions.lines=22", "-e", exe}, appArgs...)
		kittyArgs := append([]string{"--name", "ZenPomo", "--title", "ZenPomo", "-o", "initial_window_width=76c", "-o", "initial_window_height=22c", exe}, appArgs...)
		genericArgs := append([]string{"-T", "ZenPomo", "-geometry", "76x22", "-e", exe}, appArgs...)

		terminals := []struct {
			name string
			args []string
		}{
			{"gnome-terminal", gnomeArgs},
			{"alacritty", alacrittyArgs},
			{"kitty", kittyArgs},
			{"x-terminal-emulator", genericArgs},
			{"xfce4-terminal", []string{"--title=ZenPomo", "--geometry=76x22", "-e", fmt.Sprintf("%s %s", exe, strings.Join(appArgs, " "))}},
			{"konsole", append([]string{"-p", "tabtitle=ZenPomo", "-e", exe}, appArgs...)},
			{"xterm", genericArgs},
		}

		for _, t := range terminals {
			if path, err := exec.LookPath(t.name); err == nil {
				cmd = exec.Command(path, t.args...)
				break
			}
		}

		if cmd == nil {
			return fmt.Errorf("no supported terminal emulator found (install gnome-terminal)")
		}
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch TUI window: %w", err)
	}

	activeTuiCmd = cmd
	go func() {
		_ = cmd.Wait()
		tuiProcessMu.Lock()
		activeTuiCmd = nil
		tuiProcessMu.Unlock()
	}()

	return nil
}
