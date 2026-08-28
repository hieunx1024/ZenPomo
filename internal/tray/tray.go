package tray

import (
	"fmt"
	"os/signal"
	"syscall"
	"time"
	"zenpomo/internal/core"
	"zenpomo/internal/daemon"

	"fyne.io/systray"
)

// Run runs the ultra-minimal System Tray and event loop.
func Run() {
	// Ignore SIGHUP so closing any terminal never kills the tray
	signal.Ignore(syscall.SIGHUP)

	client := daemon.NewClient()
	_ = client.EnsureDaemon()

	onReady := func() {
		initialIcon := GetIconForState(core.SessionWork, core.StateStopped)
		systray.SetIcon(initialIcon)
		systray.SetTitle("ZenPomo")
		systray.SetTooltip("ZenPomo - Pomodoro Focus Timer")

		// 1. Single Clean Status Header
		mStatus := systray.AddMenuItem("ZenPomo — 25:00", "Current Pomodoro Status")
		mStatus.Disable()

		systray.AddSeparator()

		// 2. Core Actions
		mToggle := systray.AddMenuItem("Start Focus", "Start or pause timer")
		mSkip := systray.AddMenuItem("Skip Session", "Advance to next session")
		mOpenTUI := systray.AddMenuItem("Open TUI", "Open terminal interface")
		mConfig := systray.AddMenuItem("Settings", "Configure Pomodoro durations & preferences")

		systray.AddSeparator()

		// 3. Exit
		mQuit := systray.AddMenuItem("Quit", "Close tray and stop background daemon")

		lastSession := core.SessionWork
		lastState := core.StateStopped

		updateUI := func(snap core.SessionSnapshot) {
			mins := snap.RemainingSeconds / 60
			secs := snap.RemainingSeconds % 60
			timeStr := fmt.Sprintf("%02d:%02d", mins, secs)

			// Update Icon if session or running state changed
			if snap.Session != lastSession || snap.State != lastState {
				lastSession = snap.Session
				lastState = snap.State
				systray.SetIcon(GetIconForState(snap.Session, snap.State))
			}

			var stateText string
			switch snap.State {
			case core.StateRunning:
				stateText = "Running"
				mToggle.SetTitle("Pause")
			case core.StatePaused:
				stateText = "Paused"
				mToggle.SetTitle("Resume")
			default:
				stateText = "Stopped"
				mToggle.SetTitle("Start " + string(snap.Session))
			}

			// Clean Topbar Title (time only)
			systray.SetTitle(timeStr)

			// Rich Tooltip on hover
			tooltip := fmt.Sprintf("ZenPomo: %s (%s) %s\nTask: %s • Cycle %d/%d",
				snap.Session, stateText, timeStr, snap.ActiveTaskTitle, snap.CycleCount, snap.TargetCycles)
			systray.SetTooltip(tooltip)

			// Single compact header line
			mStatus.SetTitle(fmt.Sprintf("%s — %s (%s)", snap.Session, timeStr, stateText))
		}

		// Background 1-second refresh loop
		go func() {
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				snap, err := client.GetSnapshot()
				if err != nil {
					// If daemon went down, ensure it is revived
					_ = client.EnsureDaemon()
					systray.SetTitle("Offline")
					systray.SetTooltip("ZenPomo: Daemon Connecting...")
					mStatus.SetTitle("ZenPomo: Connecting...")
					continue
				}
				updateUI(snap)
			}
		}()

		// Instant responsive click dispatcher
		go func() {
			for {
				select {
				case <-mToggle.ClickedCh:
					_ = client.EnsureDaemon()
					resp, err := client.SendCommand(daemon.CmdToggle)
					if err == nil {
						updateUI(resp.Snapshot)
					}
				case <-mSkip.ClickedCh:
					_ = client.EnsureDaemon()
					resp, err := client.SendCommand(daemon.CmdSkip)
					if err == nil {
						updateUI(resp.Snapshot)
					}
				case <-mOpenTUI.ClickedCh:
					_ = client.EnsureDaemon()
					_ = FocusOrLaunchTUI()
				case <-mConfig.ClickedCh:
					_ = client.EnsureDaemon()
					_ = FocusOrLaunchConfig()
				case <-mQuit.ClickedCh:
					_, _ = client.SendCommand(daemon.CmdStop)
					systray.Quit()
					return
				}
			}
		}()
	}

	onExit := func() {
		// Clean exit
	}

	systray.Run(onReady, onExit)
}
