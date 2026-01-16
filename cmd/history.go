package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"spotiflac/backend"

	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Manage download history",
	Long: `View, search, and manage your download history.

Track all previous downloads and their details.

Examples:
  spotflac history list
  spotflac history list --limit 20
  spotflac history search "Taylor Swift"
  spotflac history clear`,
}

var historyListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show download history",
	RunE: func(cmd *cobra.Command, args []string) error {
		items, err := backend.GetHistoryItems("SpotiFLAC")
		if err != nil {
			return fmt.Errorf("failed to load history: %w", err)
		}

		if len(items) == 0 {
			fmt.Println("📭 No download history")
			return nil
		}

		fmt.Printf("📋 Download History (%d items):\n\n", len(items))

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "Title\tArtist\tAlbum\tQuality\tDate")
		fmt.Fprintln(w, "─────────────────────────────────────────────────────")

		for _, item := range items {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t\n", item.Title, item.Artists, item.Album, item.Quality)
		}
		w.Flush()

		return nil
	},
}

var historySearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search download history",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.ToLower(args[0])
		items, err := backend.GetHistoryItems("SpotiFLAC")
		if err != nil {
			return fmt.Errorf("failed to load history: %w", err)
		}

		var matching []backend.HistoryItem
		for _, item := range items {
			if strings.Contains(strings.ToLower(item.Title), query) ||
				strings.Contains(strings.ToLower(item.Artists), query) ||
				strings.Contains(strings.ToLower(item.Album), query) {
				matching = append(matching, item)
			}
		}

		if len(matching) == 0 {
			fmt.Printf("❌ No results for: %s\n", query)
			return nil
		}

		fmt.Printf("🔍 Found %d items matching '%s':\n\n", len(matching), query)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "Title\tArtist\tAlbum\tQuality")
		fmt.Fprintln(w, "──────────────────────────────────────────")

		for _, item := range matching {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", item.Title, item.Artists, item.Album, item.Quality)
		}
		w.Flush()

		return nil
	},
}

var historyExportCmd = &cobra.Command{
	Use:   "export <filepath>",
	Short: "Export history to JSON file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filepath := args[0]
		items, err := backend.GetHistoryItems("SpotiFLAC")
		if err != nil {
			return fmt.Errorf("failed to load history: %w", err)
		}

		data, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to encode history: %w", err)
		}

		if err := os.WriteFile(filepath, data, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}

		fmt.Printf("✅ Exported %d items to: %s\n", len(items), filepath)
		return nil
	},
}

var historyClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all download history",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print("⚠️  This will delete all download history. Continue? [y/N]: ")
		var response string
		fmt.Scanln(&response)

		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("❌ Cancelled")
			return nil
		}

		if err := backend.ClearHistory("SpotiFLAC"); err != nil {
			return fmt.Errorf("failed to clear history: %w", err)
		}

		fmt.Println("✅ History cleared")
		return nil
	},
}

func init() {
	historyCmd.AddCommand(historyListCmd)
	historyCmd.AddCommand(historySearchCmd)
	historyCmd.AddCommand(historyExportCmd)
	historyCmd.AddCommand(historyClearCmd)
}
