package cmd

import (
	"fmt"
	"os"
	"zenpomo/internal/tray"

	"github.com/spf13/cobra"
)

var toggleCmd = &cobra.Command{
	Use:   "toggle",
	Short: "Toggle the floating TUI window (ideal for Global OS Hotkeys)",
	Run: func(cmd *cobra.Command, args []string) {
		if err := tray.LaunchOrToggleTUI(); err != nil {
			fmt.Fprintf(os.Stderr, "Error toggling TUI window: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(toggleCmd)
}
