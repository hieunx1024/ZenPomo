package tray

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"zenpomo/internal/core"
	"zenpomo/internal/daemon"
	"zenpomo/internal/notify"
)

var (
	tuiProcessMu sync.Mutex
	activeTuiCmd *exec.Cmd
)

// FocusOrLaunchTUI brings the existing TUI window to the front or launches a new instance.
func FocusOrLaunchTUI() error {
	tuiProcessMu.Lock()
	defer tuiProcessMu.Unlock()

	client := daemon.NewClient()
	_ = client.EnsureDaemon()

	// 1. Notify running TUI to switch to Tab 1 (Timer)
	_, _ = client.SendCommand(daemon.CmdRequestTimer)

	// 2. If a TUI is already open (it heartbeats every ~250ms while running), don't spawn a
	// duplicate window — just try to bring the existing one to front. It will also self-switch
	// to the Timer tab on its next tick, picking up the pending-mode request sent above.
	if isActive, err := client.SendCommand(daemon.CmdIsTUIActive); err == nil && isActive.Success {
		notifyIfCannotFocus()
		return nil
	}

	// 3. No TUI running: try focusing (covers any stray/detached window) then launch a fresh one.
	if focusExistingWindow() {
		return nil
	}

	return launchNewTUI("tui")
}

// notifyIfCannotFocus tries to bring an already-open TUI window to the front, and — since this
// system has no wmctrl/xdotool (or the window manager otherwise blocks focus-stealing from a
// background process, common on Wayland) — falls back to a desktop notification so clicking
// "Open TUI" never silently does nothing when a window is, in fact, already open somewhere.
func notifyIfCannotFocus() {
	if focusExistingWindow() {
		return
	}
	notify.Send("ZenPomo", "ZenPomo is already open — check your other windows or workspaces.")
}

// FocusOrLaunchConfig switches the existing TUI to Settings or launches a new instance.
func FocusOrLaunchConfig() error {
	tuiProcessMu.Lock()
	defer tuiProcessMu.Unlock()

	client := daemon.NewClient()
	_ = client.EnsureDaemon()

	// 1. Notify running TUI to switch to Tab 4 (Settings)
	_, _ = client.SendCommand(daemon.CmdRequestConfig)

	// 2. Skip spawning a duplicate window if a TUI is already open; see FocusOrLaunchTUI above.
	if isActive, err := client.SendCommand(daemon.CmdIsTUIActive); err == nil && isActive.Success {
		notifyIfCannotFocus()
		return nil
	}

	// 3. No TUI running: try focusing then launch a fresh one in config mode.
	if focusExistingWindow() {
		return nil
	}

	return launchNewTUI("config")
}

// LaunchOrToggleTUI toggles (opens if closed, closes if open) the floating TUI window.
func LaunchOrToggleTUI() error {
	tuiProcessMu.Lock()
	defer tuiProcessMu.Unlock()

	// If currently active child process is running, close it (toggle behavior)
	if activeTuiCmd != nil && activeTuiCmd.Process != nil {
		_ = activeTuiCmd.Process.Signal(syscall.SIGTERM)
		activeTuiCmd = nil
		return nil
	}

	client := daemon.NewClient()
	_ = client.EnsureDaemon()

	// A TUI opened outside our own tracking (e.g. run manually, or its terminal client process
	// already exited even though the window persists) still counts as "open" — don't spawn a
	// second one on top of it just because activeTuiCmd lost track of it.
	if isActive, err := client.SendCommand(daemon.CmdIsTUIActive); err == nil && isActive.Success {
		notifyIfCannotFocus()
		return nil
	}

	if focusExistingWindow() {
		return nil
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
	// 1. Try wmctrl (X11 & XWayland)
	if p, err := exec.LookPath("wmctrl"); err == nil {
		if exec.Command(p, "-x", "-a", "ZenPomo").Run() == nil {
			return true
		}
		if exec.Command(p, "-a", "ZenPomo").Run() == nil {
			return true
		}
		if exec.Command(p, "-a", "zenpomo").Run() == nil {
			return true
		}
	}

	// 2. Try xdotool
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

var launchTUIFn = defaultLaunchTUI

func launchNewTUI(appArgs ...string) error {
	return launchTUIFn(appArgs...)
}

func defaultLaunchTUI(appArgs ...string) error {
	exe := core.GetExecutable()

	if len(appArgs) == 0 {
		appArgs = []string{"tui"}
	}

	if runtime.GOOS == "windows" {
		var cmd *exec.Cmd
		winArgs := append([]string{"--title", "ZenPomo", "-w", "0", "new-tab", exe}, appArgs...)
		if _, err := exec.LookPath("wt.exe"); err == nil {
			cmd = exec.Command("wt.exe", winArgs...)
		} else {
			cmd = exec.Command("cmd.exe", "/c", "start", "ZenPomo", exe, strings.Join(appArgs, " "))
		}
		cmd.Env = os.Environ()
		if err := cmd.Start(); err != nil {
			return err
		}
		go func() { _ = cmd.Wait() }()
		return nil
	}

	// Linux / BSD Terminal Launchers
	gnomeArgs := append([]string{"--window", "--title=ZenPomo", "--", exe}, appArgs...)
	alacrittyArgs := append([]string{"--title", "ZenPomo", "-e", exe}, appArgs...)
	kittyArgs := append([]string{"--title", "ZenPomo", exe}, appArgs...)
	genericArgs := append([]string{"-T", "ZenPomo", "-e", exe}, appArgs...)

	terminals := []struct {
		name string
		args []string
	}{
		{"gnome-terminal", gnomeArgs},
		{"kgx", append([]string{"-e", exe}, appArgs...)},
		{"ptyxis", append([]string{"--new-window", "--", exe}, appArgs...)},
		{"alacritty", alacrittyArgs},
		{"kitty", kittyArgs},
		{"x-terminal-emulator", genericArgs},
		{"xfce4-terminal", []string{"--title=ZenPomo", "-e", fmt.Sprintf("%s %s", exe, strings.Join(appArgs, " "))}},
		{"konsole", append([]string{"--new-tab", "-p", "tabtitle=ZenPomo", "-e", exe}, appArgs...)},
		{"xterm", genericArgs},
	}

	for _, t := range terminals {
		if path, err := exec.LookPath(t.name); err == nil {
			cmd := exec.Command(path, t.args...)
			cmd.Env = os.Environ()
			if err := cmd.Start(); err == nil {
				activeTuiCmd = cmd
				go func() {
					_ = cmd.Wait()
					tuiProcessMu.Lock()
					activeTuiCmd = nil
					tuiProcessMu.Unlock()
				}()
				return nil
			}
		}
	}

	// Fallback to gtk-launch or gio launch
	if p, err := exec.LookPath("gtk-launch"); err == nil {
		cmd := exec.Command(p, "zenpomo")
		cmd.Env = os.Environ()
		if err := cmd.Start(); err == nil {
			go func() { _ = cmd.Wait() }()
			return nil
		}
	}
	if p, err := exec.LookPath("gio"); err == nil {
		home, _ := os.UserHomeDir()
		desktopFile := filepath.Join(home, ".local", "share", "applications", "zenpomo.desktop")
		cmd := exec.Command(p, "launch", desktopFile)
		cmd.Env = os.Environ()
		if err := cmd.Start(); err == nil {
			go func() { _ = cmd.Wait() }()
			return nil
		}
	}

	return fmt.Errorf("no supported terminal emulator found")
}
