package cmd

import (
	"fmt"
	"os"
	"zenpomo/internal/daemon"

	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run the core background daemon process",
	Run: func(cmd *cobra.Command, args []string) {
		srv, err := daemon.NewServer()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating daemon: %v\n", err)
			os.Exit(1)
		}
		if err := srv.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Daemon stopped with error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}
