package desktop

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"zenpomo/assets"
)

// Install integrates ZenPomo into the desktop environment (Applications menu, icon, autostart).
func Install() error {
	if runtime.GOOS == "windows" {
		return installWindows()
	}
	return installLinux()
}

// EnableAutostart enables background system tray startup on login.
func EnableAutostart() error {
	if runtime.GOOS == "windows" {
		return enableAutostartWindows()
	}
	return enableAutostartLinux()
}

// DisableAutostart disables background system tray startup on login.
func DisableAutostart() error {
	if runtime.GOOS == "windows" {
		return disableAutostartWindows()
	}
	return disableAutostartLinux()
}

// IsAutostartEnabled checks if autostart is currently enabled.
func IsAutostartEnabled() bool {
	if runtime.GOOS == "windows" {
		return isAutostartWindows()
	}
	return isAutostartLinux()
}

func installLinux() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not find user home directory: %w", err)
	}

	binDir := filepath.Join(home, ".local", "bin")
	appDir := filepath.Join(home, ".local", "share", "applications")
	autostartDir := filepath.Join(home, ".config", "autostart")
	iconDir := filepath.Join(home, ".local", "share", "icons", "hicolor", "512x512", "apps")
	pixmapsDir := filepath.Join(home, ".local", "share", "pixmaps")

	for _, dir := range []string{binDir, appDir, autostartDir, iconDir, pixmapsDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// 1. Copy binary to ~/.local/bin/zenpomo
	currExe, err := os.Executable()
	if err != nil {
		currExe = "zenpomo"
	}
	targetExe := filepath.Join(binDir, "zenpomo")

	if currExe != targetExe {
		src, err := os.Open(currExe)
		if err == nil {
			defer src.Close()
			tmpTarget := targetExe + ".tmp"
			dst, err := os.OpenFile(tmpTarget, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return fmt.Errorf("failed to write binary to %s: %w", tmpTarget, err)
			}
			if _, err := io.Copy(dst, src); err != nil {
				dst.Close()
				_ = os.Remove(tmpTarget)
				return fmt.Errorf("failed to copy binary: %w", err)
			}
			dst.Close()
			_ = os.Chmod(tmpTarget, 0755)
			if err := os.Rename(tmpTarget, targetExe); err != nil {
				// Fallback if rename fails
				_ = os.Remove(targetExe)
				if err := os.Rename(tmpTarget, targetExe); err != nil {
					return fmt.Errorf("failed to replace binary %s: %w", targetExe, err)
				}
			}
		}
	} else {
		_ = os.Chmod(targetExe, 0755)
	}

	// 2. Install icon
	iconPath := filepath.Join(iconDir, "zenpomo.png")
	if len(assets.IconTomato) > 0 {
		_ = os.WriteFile(iconPath, assets.IconTomato, 0644)
		_ = os.WriteFile(filepath.Join(pixmapsDir, "zenpomo.png"), assets.IconTomato, 0644)
	}

	// 3. Create .desktop file
	desktopContent := fmt.Sprintf(`[Desktop Entry]
Name=ZenPomo
GenericName=Pomodoro Timer
Comment=Tactile Pomodoro TUI, System Tray & Widget
Exec=%s
Icon=%s
Terminal=true
Type=Application
Categories=Utility;Clock;ProjectManagement;
Keywords=pomodoro;timer;focus;productivity;
StartupNotify=true
Actions=tray;toggle;

[Desktop Action tray]
Name=Start System Tray
Exec=%s tray

[Desktop Action toggle]
Name=Toggle Window
Exec=%s toggle
`, targetExe, iconPath, targetExe, targetExe)

	desktopPath := filepath.Join(appDir, "zenpomo.desktop")
	if err := os.WriteFile(desktopPath, []byte(desktopContent), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", desktopPath, err)
	}

	// 4. Enable autostart by default
	if err := enableAutostartLinux(); err != nil {
		return err
	}

	// 5. Update desktop database if tool is present
	if p, err := exec.LookPath("update-desktop-database"); err == nil {
		_ = exec.Command(p, appDir).Run()
	}

	return nil
}

func enableAutostartLinux() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	binDir := filepath.Join(home, ".local", "bin")
	targetExe := filepath.Join(binDir, "zenpomo")
	if _, err := os.Stat(targetExe); os.IsNotExist(err) {
		if exe, err := os.Executable(); err == nil {
			targetExe = exe
		}
	}

	autostartDir := filepath.Join(home, ".config", "autostart")
	_ = os.MkdirAll(autostartDir, 0755)

	iconPath := filepath.Join(home, ".local", "share", "icons", "hicolor", "512x512", "apps", "zenpomo.png")

	autostartContent := fmt.Sprintf(`[Desktop Entry]
Name=ZenPomo System Tray
GenericName=Pomodoro Timer
Comment=ZenPomo Background System Tray Monitor
Exec=%s tray
Icon=%s
Terminal=false
Type=Application
Categories=Utility;Clock;
X-GNOME-Autostart-enabled=true
Hidden=false
NoDisplay=false
`, targetExe, iconPath)

	autostartFile := filepath.Join(autostartDir, "zenpomo-tray.desktop")
	return os.WriteFile(autostartFile, []byte(autostartContent), 0644)
}

func disableAutostartLinux() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	autostartFile := filepath.Join(home, ".config", "autostart", "zenpomo-tray.desktop")
	if err := os.Remove(autostartFile); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func isAutostartLinux() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	autostartFile := filepath.Join(home, ".config", "autostart", "zenpomo-tray.desktop")
	_, err = os.Stat(autostartFile)
	return err == nil
}

func installWindows() error {
	return enableAutostartWindows()
}

func enableAutostartWindows() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command("reg", "add", `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, "/v", "ZenPomoTray", "/t", "REG_SZ", "/d", fmt.Sprintf("\"%s\" tray", exe), "/f")
	return cmd.Run()
}

func disableAutostartWindows() error {
	cmd := exec.Command("reg", "delete", `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, "/v", "ZenPomoTray", "/f")
	return cmd.Run()
}

func isAutostartWindows() bool {
	cmd := exec.Command("reg", "query", `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, "/v", "ZenPomoTray")
	return cmd.Run() == nil
}
