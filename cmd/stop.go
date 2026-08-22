package cmd

import (
	"fmt"
	"os"
	"zenpomo/internal/daemon"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the background daemon process",
	Run: func(cmd *cobra.Command, args []string) {
		client := daemon.NewClient()
		if !client.IsRunning() {
			fmt.Println("ZenPomo daemon is not running.")
			return
		}
		_, err := client.SendCommand(daemon.CmdStop)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error stopping daemon: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ ZenPomo daemon stopped successfully.")
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
