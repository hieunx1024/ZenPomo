# ZenPomo

**English** | [Vietnamese](README.vn.md)

ZenPomo is a lightweight, distraction-free Pomodoro timer featuring a tactile Terminal User Interface (TUI), native System Tray integration, and a background daemon. Built with Go for **Linux** (Ubuntu, Debian, Fedora, Arch) and **Windows 10/11**.

Consumes minimal system resources (<15MB RAM), requires zero external runtime dependencies (100% pure Go static binary, no CGO), and is fully keyboard-driven.

---

## Features

- **Tactile TUI:** Minimalist terminal interface with large ASCII block digits, smooth rendering, and Vim-style key navigation.
- **Native System Tray:** Lives on the Ubuntu Topbar (GNOME, KDE, Waybar) and Windows Taskbar with dynamic state icons and a quick control menu.
- **Background Daemon:** The timer runs independently in the background; opening or closing the TUI window never interrupts your active focus session.
- **Seamless Auto-flow:** Automatically transitions between Work and Break sessions with customizable intervals.
- **Top Bar & Widget Integration:** Built-in `zenpomo status` command supports both plain text and JSON outputs for **Waybar**, **Polybar**, **Tmux**, **i3/Sway**, or custom scripts.
- **Offline Audio & Desktop Notifications:** Embedded high-quality audio cues (zero latency, offline) and native desktop notifications on session completion.
- **Zero-Config & Pure Go:** Built with `CGO_ENABLED=0` without requiring C/C++ development headers (`libappindicator-dev` or `gcc`).

---

## Downloads & Installation

Pre-built binaries are available on the **[Releases](https://github.com/hieunx1024/ZenPomo/releases)** page.

### 1. Windows 10/11
* **Windows Setup Wizard (.exe)** *(Recommended)*:  
  Download `zenpomo_1.0.0_windows_setup.exe` $\rightarrow$ Double-click to install. Automatically creates Desktop and Start Menu shortcuts, configures background tray startup, and registers with Windows Add/Remove Programs.
* **Portable Binary (.exe)**:  
  Download `zenpomo_1.0.0_windows_amd64.exe` $\rightarrow$ Rename to `zenpomo.exe` and run directly.

### 2. Ubuntu / Debian (.deb)
Download `zenpomo_1.0.0_linux_amd64.deb` and double-click to install, or run:
```bash
sudo dpkg -i zenpomo_1.0.0_linux_amd64.deb
```
> The `.deb` package automatically adds ZenPomo to your Applications Menu and enables background System Tray autostart on login.

### 3. Generic Linux (Standalone Binary)
Download `zenpomo_1.0.0_linux_amd64`:
```bash
chmod +x zenpomo_1.0.0_linux_amd64
sudo mv zenpomo_1.0.0_linux_amd64 /usr/local/bin/zenpomo
zenpomo install
```

---

## Building from Source

Requires **Go 1.22+**.

```bash
# Clone the repository
git clone https://github.com/hieunx1024/ZenPomo.git
cd ZenPomo

# Build and install to local environment
make install
```

Or build specific targets manually:
```bash
make build-linux      # Build Linux binary (bin/zenpomo)
make build-windows    # Build Windows executable (bin/zenpomo.exe)
make deb              # Build Debian package (dist/zenpomo_1.0.0_amd64.deb)
make build-all        # Build all targets
```

---

## Usage

### Launching the Interface

```bash
# Open full-screen TUI
zenpomo

# Open directly into Configuration Settings modal
zenpomo config
```

### Command Line Interface (CLI)

```bash
zenpomo start       # Start the timer
zenpomo pause       # Pause the timer
zenpomo skip        # Advance to the next session
zenpomo reset       # Reset current session
zenpomo stop        # Stop background daemon completely
zenpomo toggle      # Toggle (show/hide) the TUI window
```

### Waybar / Polybar / Tmux Integration

Use `zenpomo status` to output live countdown data:

```bash
# Plain text output: [24:35] [Work: Running] | Task: General Focus
zenpomo status

# JSON output (for Waybar / custom status bar scripts)
zenpomo status --format json
```

Example Waybar configuration (`~/.config/waybar/config`):

```json
"custom/zenpomo": {
    "exec": "zenpomo status --format json",
    "return-type": "json",
    "interval": 1,
    "on-click": "zenpomo toggle"
}
```

### Managing System Tray Autostart

```bash
zenpomo autostart enable     # Enable startup on login
zenpomo autostart disable    # Disable startup
zenpomo autostart status     # Check autostart status
```

---

## Keybindings

| Key | Action |
| :--- | :--- |
| `Space` | Start / Pause timer |
| `n` | Skip / Advance to next session |
| `r` | Reset current session timer |
| `c` | Open Configuration Settings modal |
| `a` | Add new task to todo list |
| `d` | Toggle task completion (Done / Undone) |
| `x` | Delete selected task |
| `Enter` | Set selected task as active timer task |
| `j` / `↓` | Move selection down |
| `k` / `↑` | Move selection up |
| `m` | Toggle audio cues (Mute / Unmute) |
| `q` / `Ctrl+C` | Close TUI (daemon and tray continue running) |

---

## Global Hotkey (Ubuntu / GNOME)

To toggle ZenPomo instantly with a system-wide hotkey:

1. Go to **Settings** $\rightarrow$ **Keyboard** $\rightarrow$ **Keyboard Shortcuts** $\rightarrow$ **Custom Shortcuts**.
2. Add a new shortcut:
   - **Name:** `ZenPomo Toggle`
   - **Command:** `zenpomo toggle`
   - **Shortcut:** `Ctrl + Alt + P` (or your preferred key combination).

---

## License

This project is licensed under the [MIT License](LICENSE).
