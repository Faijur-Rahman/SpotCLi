package cmd

import (
	"fmt"
	"path/filepath"

	"spotiflac/backend"

	"github.com/spf13/cobra"
)

var convertCmd = &cobra.Command{
	Use:   "convert <input-file...>",
	Short: "Convert audio files between formats",
	Long: `Convert audio files to different formats and bitrates.

Supports conversion to MP3, M4A, OGG, and other formats.

Examples:
  spotflac convert song.flac --format mp3
  spotflac convert song.flac --format mp3 --bitrate 320k
  spotflac convert file1.flac file2.flac --format ogg --output ~/converted`,
	Args: cobra.MinimumNArgs(1),
	RunE: runConvert,
}

var (
	convertFormat  string
	convertBitrate string
	convertCodec   string
	convertOutput  string
)

func init() {
	convertCmd.Flags().StringVarP(&convertFormat, "format", "f", "mp3", "Output format: mp3, m4a, ogg, opus, flac")
	convertCmd.Flags().StringVarP(&convertBitrate, "bitrate", "b", "320k", "Bitrate: 128k, 192k, 256k, 320k, etc.")
	convertCmd.Flags().StringVar(&convertCodec, "codec", "", "Codec override (optional)")
	convertCmd.Flags().StringVarP(&convertOutput, "output", "o", "", "Output directory (default: same as input)")
}

func runConvert(cmd *cobra.Command, args []string) error {
	inputFiles := args

	// Validate format
	validFormats := map[string]bool{
		"mp3":  true,
		"m4a":  true,
		"ogg":  true,
		"opus": true,
		"flac": true,
	}
	if !validFormats[convertFormat] {
		return fmt.Errorf("unsupported format: %s", convertFormat)
	}

	fmt.Printf("🎵 Converting %d file(s) to %s...\n\n", len(inputFiles), convertFormat)

	req := backend.ConvertAudioRequest{
		InputFiles:   inputFiles,
		OutputFormat: convertFormat,
		Bitrate:      convertBitrate,
		Codec:        convertCodec,
	}

	results, err := backend.ConvertAudio(req)
	if err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	successCount := 0
	for i, result := range results {
		if result.Success {
			successCount++
			fmt.Printf("✅ %d. %s → %s\n", i+1, filepath.Base(inputFiles[i]), result.OutputFile)
		} else {
			fmt.Printf("❌ %d. %s - Error: %s\n", i+1, filepath.Base(inputFiles[i]), result.Error)
		}
	}

	fmt.Printf("\n📊 Summary: %d/%d files converted successfully\n", successCount, len(inputFiles))
	return nil
}

var lyricsCmd = &cobra.Command{
	Use:   "lyrics",
	Short: "Download and manage lyrics",
	Long:  `Manage song lyrics - download, embed, and extract.`,
}

var lyricsDownloadCmd = &cobra.Command{
	Use:   "download <spotify-id>",
	Short: "Download lyrics for a track",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		spotifyID := args[0]

		fmt.Printf("📥 Downloading lyrics for: %s\n", spotifyID)

		client := backend.NewLyricsClient()
		lyricsResp, source, err := client.FetchLyricsAllSources(spotifyID, "", "", 0)
		if err != nil {
			return fmt.Errorf("failed to fetch lyrics: %w", err)
		}

		if lyricsResp == nil || len(lyricsResp.Lines) == 0 {
			return fmt.Errorf("no lyrics found")
		}

		fmt.Printf("✅ Found %d lines from: %s\n", len(lyricsResp.Lines), source)
		fmt.Printf("Sync type: %s\n\n", lyricsResp.SyncType)

		for _, line := range lyricsResp.Lines {
			fmt.Println(line)
		}

		return nil
	},
}

var coverCmd = &cobra.Command{
	Use:   "cover",
	Short: "Download and manage cover art",
	Long:  `Download cover art and album artwork.`,
}

var coverDownloadCmd = &cobra.Command{
	Use:   "download <url>",
	Short: "Download cover from URL",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]
		outputDir := "."

		fmt.Printf("📥 Downloading cover from: %s\n", url)

		req := backend.CoverDownloadRequest{
			CoverURL:  url,
			OutputDir: outputDir,
		}

		client := backend.NewCoverClient()
		resp, err := client.DownloadCover(req)
		if err != nil {
			return fmt.Errorf("download failed: %w", err)
		}

		if !resp.Success {
			return fmt.Errorf("download failed: %s", resp.Error)
		}

		fmt.Printf("✅ %s\n", resp.Message)
		if resp.File != "" {
			fmt.Printf("📁 Saved to: %s\n", resp.File)
		}

		return nil
	},
}

var availabilityCmd = &cobra.Command{
	Use:   "availability <spotify-id>",
	Short: "Check track availability on streaming services",
	Long: `Check which streaming services have a track available.

Shows availability status on Tidal, Qobuz, Amazon Music, and other platforms.

Examples:
  spotflac availability 4cOdK2wGLETKBW3PvgPWqLv
  spotflac availability 4cOdK2wGLETKBW3PvgPWqLv --isrc USRC17607839`,
	Args: cobra.ExactArgs(1),
	RunE: runAvailability,
}

var (
	availabilityISRC string
)

func init() {
	availabilityCmd.Flags().StringVar(&availabilityISRC, "isrc", "", "Optional ISRC code")

	lyricsCmd.AddCommand(lyricsDownloadCmd)
	coverCmd.AddCommand(coverDownloadCmd)
}

func runAvailability(cmd *cobra.Command, args []string) error {
	spotifyID := args[0]

	fmt.Printf("🔍 Checking availability for: %s\n", spotifyID)

	client := backend.NewSongLinkClient()
	availability, err := client.CheckTrackAvailability(spotifyID, availabilityISRC)
	if err != nil {
		return fmt.Errorf("failed to check availability: %w", err)
	}

	printAvailability(availability)
	return nil
}

func printAvailability(data interface{}) {
	if mapData, ok := data.(map[string]interface{}); ok {
		fmt.Println("\n📡 Streaming Service Availability")
		fmt.Println("════════════════════════════════════════════")

		services := []string{"spotify", "tidal", "qobuz", "amazonmusic", "applemusic", "deezer", "youtube"}

		for _, service := range services {
			available := false
			if val, exists := mapData[service]; exists && val != nil {
				available = true
			}

			status := "❌ Not available"
			if available {
				status = "✅ Available"
			}

			fmt.Printf("%-15s %s\n", formatServiceName(service), status)
		}
	}
}

func formatServiceName(service string) string {
	names := map[string]string{
		"spotify":     "Spotify",
		"tidal":       "Tidal",
		"qobuz":       "Qobuz",
		"amazonmusic": "Amazon Music",
		"applemusic":  "Apple Music",
		"deezer":      "Deezer",
		"youtube":     "YouTube Music",
	}
	if name, ok := names[service]; ok {
		return name
	}
	return service
}
