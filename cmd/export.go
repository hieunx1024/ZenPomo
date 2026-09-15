package cmd

import (
	"fmt"
	"os"
	"zenpomo/internal/analytics"
	"zenpomo/internal/storage"

	"github.com/spf13/cobra"
)

var exportFormat string
var exportOutput string

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export focus history and tasks (csv, md, or json)",
	Long: `Export your ZenPomo productivity history without opening the TUI.

Formats:
  csv   Daily focus stats as a spreadsheet-friendly CSV (date,focus_minutes,completed_pomos,completed_tasks)
  md    Full Markdown report with overview metrics, weekly/monthly breakdown, and task list
  json  Raw backup of the entire local database (config, tasks, and daily stats)`,
	Run: func(cmd *cobra.Command, args []string) {
		store, err := storage.NewStore()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open local data store: %v\n", err)
			os.Exit(1)
		}

		todayStats := store.GetTodayStats()
		allStats := store.GetAllStats()
		tasks := store.GetTasks()

		out := exportOutput
		var exportErr error

		switch exportFormat {
		case "csv":
			if out == "" {
				out = "zenpomo-stats.csv"
			}
			exportErr = analytics.ExportCSV(out, allStats, todayStats)
		case "md", "markdown":
			if out == "" {
				out = "zenpomo-stats.md"
			}
			summary := analytics.ComputeSummary(allStats, todayStats)
			exportErr = analytics.ExportMarkdown(out, summary, tasks, allStats, todayStats)
		case "json":
			if out == "" {
				out = "zenpomo-backup.json"
			}
			exportErr = store.ExportJSON(out)
		default:
			fmt.Fprintf(os.Stderr, "Error: unknown format %q (expected csv, md, or json)\n", exportFormat)
			os.Exit(1)
		}

		if exportErr != nil {
			fmt.Fprintf(os.Stderr, "Error: export failed: %v\n", exportErr)
			os.Exit(1)
		}
		fmt.Printf("✔ Exported %s data to %s\n", exportFormat, out)
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "csv", "Export format: csv, md, or json")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file path (defaults to zenpomo-stats.<ext> or zenpomo-backup.json)")
	rootCmd.AddCommand(exportCmd)
}
