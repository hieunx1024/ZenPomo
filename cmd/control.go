package cmd

import (
	"fmt"
	"os"
	"zenpomo/internal/daemon"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start or resume countdown",
	Run: func(cmd *cobra.Command, args []string) {
		client := daemon.NewClient()
		_ = client.EnsureDaemon()
		resp, err := client.SendCommand(daemon.CmdStart)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("▶ Timer started (%s - %02d:%02d)\n", resp.Snapshot.Session, resp.Snapshot.RemainingSeconds/60, resp.Snapshot.RemainingSeconds%60)
	},
}

var pauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause countdown",
	Run: func(cmd *cobra.Command, args []string) {
		client := daemon.NewClient()
		resp, err := client.SendCommand(daemon.CmdPause)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("⏸ Timer paused (%02d:%02d remaining)\n", resp.Snapshot.RemainingSeconds/60, resp.Snapshot.RemainingSeconds%60)
	},
}

var skipCmd = &cobra.Command{
	Use:   "skip",
	Short: "Skip to next session",
	Run: func(cmd *cobra.Command, args []string) {
		client := daemon.NewClient()
		_ = client.EnsureDaemon()
		resp, err := client.SendCommand(daemon.CmdSkip)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("⏭ Advanced to: %s (%02d:%02d)\n", resp.Snapshot.Session, resp.Snapshot.RemainingSeconds/60, resp.Snapshot.RemainingSeconds%60)
	},
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset current session timer",
	Run: func(cmd *cobra.Command, args []string) {
		client := daemon.NewClient()
		resp, err := client.SendCommand(daemon.CmdReset)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("↺ Timer reset to %02d:%02d\n", resp.Snapshot.RemainingSeconds/60, resp.Snapshot.RemainingSeconds%60)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(pauseCmd)
	rootCmd.AddCommand(skipCmd)
	rootCmd.AddCommand(resetCmd)
}
