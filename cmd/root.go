package cmd

import (
	"fmt"
	"os"
	"zenpomo/internal/tray"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "zenpomo",
	Short: "ZenPomo - Tactile Pomodoro TUI, System Tray & Widget",
	Long: `ZenPomo is a lightweight, distraction-free Pomodoro timer built with Go.
It features a tactile Vim-navigated TUI, background daemon, System Tray integration,
embedded local audio cues, and instant status outputs for Waybar, Tmux, and Taskbars.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure system tray is running in background
		tray.EnsureTray()
		// Default action: launch TUI
		runTUI()
	},
}

// Execute runs the root CLI command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
