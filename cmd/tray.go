package cmd

import (
	"zenpomo/internal/tray"

	"github.com/spf13/cobra"
)

var trayCmd = &cobra.Command{
	Use:   "tray",
	Short: "Run the System Tray icon and background event loop",
	Run: func(cmd *cobra.Command, args []string) {
		tray.Run()
	},
}

func init() {
	rootCmd.AddCommand(trayCmd)
}
