package cmd

import (
	"fmt"
	"os"
	"zenpomo/internal/desktop"

	"github.com/spf13/cobra"
)

var autostartCmd = &cobra.Command{
	Use:   "autostart [enable|disable|status]",
	Short: "Manage automatic background System Tray startup on login",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		action := "status"
		if len(args) > 0 {
			action = args[0]
		}

		switch action {
		case "enable", "on":
			if err := desktop.EnableAutostart(); err != nil {
				fmt.Fprintf(os.Stderr, "Error enabling autostart: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✓ ZenPomo System Tray autostart enabled on login.")
		case "disable", "off":
			if err := desktop.DisableAutostart(); err != nil {
				fmt.Fprintf(os.Stderr, "Error disabling autostart: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✓ ZenPomo System Tray autostart disabled.")
		case "status":
			if desktop.IsAutostartEnabled() {
				fmt.Println("✓ ZenPomo System Tray autostart is currently [ENABLED].")
			} else {
				fmt.Println("○ ZenPomo System Tray autostart is currently [DISABLED]. (Run 'zenpomo autostart enable' to turn on)")
			}
		default:
			fmt.Println("Unknown action. Usage: zenpomo autostart [enable|disable|status]")
		}
	},
}

func init() {
	rootCmd.AddCommand(autostartCmd)
}
