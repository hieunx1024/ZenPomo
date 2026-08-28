package cmd

import (
	"fmt"
	"os"
	"zenpomo/internal/app"
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
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	tuiCmd.Flags().BoolVarP(&tuiConfigFlag, "config", "c", false, "Start directly in Configuration Settings modal")
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(configCmd)
}
