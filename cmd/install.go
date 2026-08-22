package cmd

import (
	"fmt"
	"os"
	"zenpomo/internal/daemon"
	"zenpomo/internal/desktop"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Integrate ZenPomo with Desktop (App Menu, Icon, System Tray autostart)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚀 Installing ZenPomo Desktop Integration...")

		if err := desktop.Install(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Installation failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✓ Copied binary to ~/.local/bin/zenpomo")
		fmt.Println("✓ Installed app icon to ~/.local/share/icons/")
		fmt.Println("✓ Created desktop launcher at ~/.local/share/applications/zenpomo.desktop")
		fmt.Println("✓ Enabled background System Tray autostart at ~/.config/autostart/zenpomo-tray.desktop")

		// Start daemon & tray immediately
		client := daemon.NewClient()
		_ = client.EnsureDaemon()

		fmt.Println("\n🎉 ZenPomo installed successfully! You can now find ZenPomo in your Application Menu or run 'zenpomo'.")
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
