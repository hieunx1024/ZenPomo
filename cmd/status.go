package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"zenpomo/internal/daemon"

	"github.com/spf13/cobra"
)

var (
	formatFlag string
	autoStart  bool
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Print current Pomodoro timer status (ideal for Waybar, Polybar, Tmux, Scripts)",
	Run: func(cmd *cobra.Command, args []string) {
		client := daemon.NewClient()
		if autoStart {
			_ = client.EnsureDaemon()
		}

		snap, err := client.GetSnapshot()
		if err != nil {
			if formatFlag == "json" {
				fmt.Println(`{"status": "offline", "time": "00:00"}`)
			} else {
				fmt.Println("[Offline]")
			}
			os.Exit(0)
		}

		mins := snap.RemainingSeconds / 60
		secs := snap.RemainingSeconds % 60
		timeStr := fmt.Sprintf("%02d:%02d", mins, secs)

		if formatFlag == "json" {
			data := map[string]interface{}{
				"state":             snap.State,
				"session":           snap.Session,
				"time_left":         timeStr,
				"remaining_seconds": snap.RemainingSeconds,
				"total_seconds":     snap.TotalSeconds,
				"progress":          snap.ProgressRatio,
				"active_task":       snap.ActiveTaskTitle,
				"cycle":             snap.CycleCount,
			}
			b, _ := json.Marshal(data)
			fmt.Println(string(b))
		} else {
			fmt.Printf("[%s] [%s: %s] | Task: %s\n", timeStr, snap.Session, snap.State, snap.ActiveTaskTitle)
		}
	},
}

func init() {
	statusCmd.Flags().StringVarP(&formatFlag, "format", "f", "text", "Output format: text or json")
	statusCmd.Flags().BoolVarP(&autoStart, "autostart", "a", false, "Automatically start background daemon if offline")
	rootCmd.AddCommand(statusCmd)
}
