package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "spotflac",
	Short: "SpotiFLAC - Download Spotify tracks in FLAC quality from Tidal, Qobuz & Amazon Music",
	Long: `SpotiFLAC is a command-line tool for downloading Spotify tracks in true FLAC quality
from third-party streaming services including Tidal, Qobuz, and Amazon Music.

No Spotify account required. No account needed for third-party services either.

Examples:
  spotflac download https://open.spotify.com/track/... --service tidal
  spotflac search "Taylor Swift" --type track
  spotflac config set download-path ~/Music
  spotflac history list`,
	Version: "7.0.6",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: false,
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(metadataCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(historyCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(lyricsCmd)
	rootCmd.AddCommand(coverCmd)
	rootCmd.AddCommand(availabilityCmd)
}
