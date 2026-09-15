package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"zenpomo/internal/app"
	"zenpomo/internal/daemon"
	"zenpomo/internal/tray"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	tuiConfigFlag bool
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the full-screen terminal interface",
	Run: func(cmd *cobra.Command, args []string) {
		mode := app.ModeNormal
		if tuiConfigFlag {
			mode = app.ModeConfig
		}
		runTUI(mode)
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Open ZenPomo configuration modal directly in TUI",
	Run: func(cmd *cobra.Command, args []string) {
		runTUI(app.ModeConfig)
	},
}

func runTUI(mode ...app.InputMode) {
	tray.EnsureTray()
	initialMode := app.ModeNormal
	if len(mode) > 0 {
		initialMode = mode[0]
	}
	p := tea.NewProgram(
		app.NewModel(initialMode),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// The terminal window closing delivers SIGHUP (SIGTERM if killed some other way); Go's
	// default disposition for both is to terminate immediately, which would skip the
	// disconnect notification below entirely. Catch them, quit the program gracefully instead,
	// and let the normal post-Run cleanup run.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGTERM)
	go func() {
		<-sigChan
		p.Quit()
	}()

	_, runErr := p.Run()

	// Tell the daemon this TUI is gone right away, instead of leaving it to notice via
	// heartbeat timeout (up to ~1.5s) — otherwise closing the TUI and immediately reopening it
	// from the tray can see a stale "still active" and silently skip launching a new window.
	_, _ = daemon.NewClient().SendCommand(daemon.CmdTUIDisconnect)

	if runErr != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", runErr)
		os.Exit(1)
	}
}

func init() {
	tuiCmd.Flags().BoolVarP(&tuiConfigFlag, "config", "c", false, "Start directly in Configuration Settings modal")
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(configCmd)
}
