package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"spotiflac/backend"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for tracks, albums, artists, or playlists on Spotify",
	Long: `Search Spotify for tracks, albums, artists, or playlists.

By default, searches for tracks. Use --type to search for other types.

Examples:
  spotflac search "Taylor Swift Anti-Hero"
  spotflac search "The Beatles" --type artist
  spotflac search "1989" --type album --limit 20
  spotflac search "Discover Weekly" --type playlist`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

var (
	searchType  string
	searchLimit int
)

func init() {
	searchCmd.Flags().StringVar(&searchType, "type", "track", "Search type: track, album, artist, playlist")
	searchCmd.Flags().IntVar(&searchLimit, "limit", 10, "Maximum results to return")
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	fmt.Printf("🔍 Searching %s for: %s\n", searchType, query)

	results, err := backend.SearchSpotifyByType(cmd.Context(), query, searchType, searchLimit, 0)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("❌ No results found")
		return nil
	}

	fmt.Printf("\n📋 Found %d results:\n\n", len(results))

	switch searchType {
	case "track":
		printTrackResults(results)
	case "album":
		printAlbumResults(results)
	case "artist":
		printArtistResults(results)
	case "playlist":
		printPlaylistResults(results)
	}

	return nil
}

func printTrackResults(results []backend.SearchResult) {
	for i, result := range results {
		fmt.Printf("%d. %s\n", i+1, result.Name)
		fmt.Printf("   🎤 Artist: %s\n", result.Artists)
		fmt.Printf("   💿 Album: %s\n", result.AlbumName)
		fmt.Printf("   🔗 Spotify: https://open.spotify.com/track/%s\n", result.ID)
		fmt.Printf("   ID: %s\n\n", result.ID)
	}
}

func printAlbumResults(results []backend.SearchResult) {
	for i, result := range results {
		fmt.Printf("%d. %s\n", i+1, result.Name)
		fmt.Printf("   🎤 Artist: %s\n", result.Artists)
		if result.ReleaseDate != "" {
			fmt.Printf("   📅 Release: %s\n", result.ReleaseDate)
		}
		fmt.Printf("   🔗 Spotify: https://open.spotify.com/album/%s\n", result.ID)
		fmt.Printf("   ID: %s\n\n", result.ID)
	}
}

func printArtistResults(results []backend.SearchResult) {
	for i, result := range results {
		fmt.Printf("%d. %s\n", i+1, result.Name)
		if result.TotalTracks > 0 {
			fmt.Printf("   ⭐ Popularity: %d\n", result.TotalTracks)
		}
		fmt.Printf("   🔗 Spotify: https://open.spotify.com/artist/%s\n", result.ID)
		fmt.Printf("   ID: %s\n\n", result.ID)
	}
}

func printPlaylistResults(results []backend.SearchResult) {
	for i, result := range results {
		fmt.Printf("%d. %s\n", i+1, result.Name)
		if result.AlbumName != "" {
			fmt.Printf("   📝 Description: %s\n", result.AlbumName)
		}
		fmt.Printf("   🔗 Spotify: https://open.spotify.com/playlist/%s\n", result.ID)
		fmt.Printf("   ID: %s\n\n", result.ID)
	}
}

var metadataCmd = &cobra.Command{
	Use:   "metadata <spotify-url|spotify-id>",
	Short: "Fetch and display track metadata",
	Long: `Fetch comprehensive metadata for a Spotify track.

Shows track details, artist info, album data, and more in JSON format.

Examples:
  spotflac metadata https://open.spotify.com/track/4cOdK2wGLETKBW3PvgPWqLv
  spotflac metadata 4cOdK2wGLETKBW3PvgPWqLv --format json
  spotflac metadata 4cOdK2wGLETKBW3PvgPWqLv --format pretty`,
	Args: cobra.ExactArgs(1),
	RunE: runMetadata,
}

var (
	metadataFormat  string
	metadataBatch   bool
	metadataDelay   float64
	metadataTimeout float64
)

func init() {
	metadataCmd.Flags().StringVar(&metadataFormat, "format", "pretty", "Output format: json or pretty")
	metadataCmd.Flags().BoolVar(&metadataBatch, "batch", false, "Process as batch request")
	metadataCmd.Flags().Float64Var(&metadataDelay, "delay", 1.0, "Delay between requests (seconds)")
	metadataCmd.Flags().Float64Var(&metadataTimeout, "timeout", 300.0, "Request timeout (seconds)")
}

func runMetadata(cmd *cobra.Command, args []string) error {
	spotifyInput := args[0]

	// Parse Spotify URL or ID
	spotifyURL := spotifyInput
	if !strings.Contains(spotifyInput, "spotify.com") {
		spotifyURL = fmt.Sprintf("https://open.spotify.com/track/%s", spotifyInput)
	}

	fmt.Printf("📍 Fetching metadata...\n")

	data, err := backend.GetFilteredSpotifyData(cmd.Context(), spotifyURL, metadataBatch, 0)
	if err != nil {
		return fmt.Errorf("failed to fetch metadata: %w", err)
	}

	if metadataFormat == "json" {
		jsonBytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to encode response: %w", err)
		}
		fmt.Println(string(jsonBytes))
	} else {
		printFormattedMetadata(data)
	}

	return nil
}

func printFormattedMetadata(data interface{}) {
	// Handle backend.TrackResponse struct
	if trackResp, ok := data.(backend.TrackResponse); ok {
		track := trackResp.Track
		fmt.Println("\n🎵 Track Information")
		fmt.Println("═════════════════════════════════════════")
		fmt.Printf("Name: %s\n", track.Name)
		fmt.Printf("ISRC: %s\n", track.ISRC)
		fmt.Printf("Duration: %d ms\n", track.DurationMS)
		fmt.Printf("Track Number: %d\n", track.TrackNumber)

		fmt.Println("\n🎤 Artists")
		fmt.Println("─────────────────────────────────────────")
		fmt.Printf("%s\n", track.Artists)

		fmt.Println("\n💿 Album Information")
		fmt.Println("─────────────────────────────────────────")
		fmt.Printf("Name: %s\n", track.AlbumName)
		fmt.Printf("Artist: %s\n", track.AlbumArtist)
		fmt.Printf("Release Date: %s\n", track.ReleaseDate)
		fmt.Printf("Total Tracks: %d\n", track.TotalTracks)

		fmt.Println("\n🔗 Links")
		fmt.Println("─────────────────────────────────────────")
		fmt.Printf("Spotify: %s\n", track.ExternalURL)
		fmt.Printf("Cover: %s\n", track.Images)
		return
	}

	// Also handle map-based responses for compatibility
	if mapData, ok := data.(map[string]interface{}); ok {
		if track, ok := mapData["track"].(map[string]interface{}); ok {
			fmt.Println("\n🎵 Track Information")
			fmt.Println("═════════════════════════════════════════")

			if name, ok := track["name"].(string); ok {
				fmt.Printf("Name: %s\n", name)
			}

			if isrc, ok := track["isrc"].(string); ok {
				fmt.Printf("ISRC: %s\n", isrc)
			}

			if explicit, ok := track["explicit"].(bool); ok {
				status := "No"
				if explicit {
					status = "Yes"
				}
				fmt.Printf("Explicit: %s\n", status)
			}

			if duration, ok := track["duration_ms"].(float64); ok {
				minutes := int(duration) / 60000
				seconds := (int(duration) % 60000) / 1000
				fmt.Printf("Duration: %d:%02d\n", minutes, seconds)
			}

			fmt.Println("\n🎤 Artists")
			fmt.Println("─────────────────────────────────────────")
			if artists, ok := track["artists"].([]interface{}); ok {
				for i, artist := range artists {
					if artistMap, ok := artist.(map[string]interface{}); ok {
						if name, ok := artistMap["name"].(string); ok {
							fmt.Printf("%d. %s\n", i+1, name)
						}
					}
				}
			}

			fmt.Println("\n💿 Album Information")
			fmt.Println("─────────────────────────────────────────")
			if album, ok := track["album"].(map[string]interface{}); ok {
				if name, ok := album["name"].(string); ok {
					fmt.Printf("Name: %s\n", name)
				}
				if releaseDate, ok := album["release_date"].(string); ok {
					fmt.Printf("Release Date: %s\n", releaseDate)
				}
				if totalTracks, ok := album["total_tracks"].(float64); ok {
					fmt.Printf("Total Tracks: %d\n", int(totalTracks))
				}
			}

			fmt.Println("\n🔗 Links")
			fmt.Println("─────────────────────────────────────────")
			if url, ok := track["external_urls"].(map[string]interface{}); ok {
				if spotify, ok := url["spotify"].(string); ok {
					fmt.Printf("Spotify: %s\n", spotify)
				}
			}
		}
	}
}
