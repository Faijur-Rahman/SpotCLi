package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"spotiflac/backend"

	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download <spotify-url|spotify-id>",
	Short: "Download a Spotify track in FLAC quality",
	Long: `Download a Spotify track in FLAC quality from Tidal, Qobuz, or Amazon Music.

Supports Spotify URLs (https://open.spotify.com/track/...) or Spotify track IDs.

Examples:
  spotflac download https://open.spotify.com/track/4cOdK2wGLETKBW3PvgPWqLv
  spotflac download 4cOdK2wGLETKBW3PvgPWqLv --service tidal
  spotflac download 4cOdK2wGLETKBW3PvgPWqLv -o ~/Music --embed-lyrics
  spotflac download 4cOdK2wGLETKBW3PvgPWqLv --service qobuz --quality 24`,
	Args: cobra.ExactArgs(1),
	RunE: runDownload,
}

var (
	downloadOutputDir       string
	downloadService         string
	downloadQuality         string
	downloadFormat          string
	downloadFilenameFormat  string
	downloadFolderTemplate  string
	downloadEmbedLyrics     bool
	downloadEmbedMaxQuality bool
	downloadTrackNumber     bool
	downloadUseAlbumTrack   bool
	downloadTidalAPI        string
)

func init() {
	downloadCmd.Flags().StringVarP(&downloadOutputDir, "output", "o", "", "Output directory (default: music folder)")
	downloadCmd.Flags().StringVarP(&downloadService, "service", "s", "auto", "Streaming service: auto, tidal, qobuz, amazon")
	downloadCmd.Flags().StringVarP(&downloadQuality, "quality", "q", "", "Audio quality (tidal: LOSSLESS|HI_RES_LOSSLESS, qobuz: 6|7, amazon: original)")
	downloadCmd.Flags().StringVarP(&downloadFormat, "format", "f", "LOSSLESS", "Audio format (deprecated, use --quality)")
	downloadCmd.Flags().StringVar(&downloadFilenameFormat, "filename", "title-artist", "Filename format: title|title-artist|artist-title|track-title|artist-album-title|custom")
	downloadCmd.Flags().StringVar(&downloadFolderTemplate, "folder", "none", "Folder structure: none|artist|album|artist-album|year-album|year-artist-album|custom")
	downloadCmd.Flags().BoolVar(&downloadEmbedLyrics, "embed-lyrics", false, "Embed lyrics in FLAC files")
	downloadCmd.Flags().BoolVar(&downloadEmbedMaxQuality, "embed-max-quality-cover", false, "Embed maximum quality album cover")
	downloadCmd.Flags().BoolVar(&downloadTrackNumber, "track-number", false, "Include track number in filename")
	downloadCmd.Flags().BoolVar(&downloadUseAlbumTrack, "use-album-track", false, "Use album track number instead of position")
	downloadCmd.Flags().StringVar(&downloadTidalAPI, "tidal-api", "auto", "Tidal API endpoint (auto or custom URL)")
}

func runDownload(cmd *cobra.Command, args []string) error {
	spotifyInput := args[0]

	// Parse Spotify URL or ID
	spotifyID := spotifyInput
	if strings.Contains(spotifyInput, "spotify.com") {
		spotifyID = extractSpotifyID(spotifyInput)
		if spotifyID == "" {
			return fmt.Errorf("invalid Spotify URL")
		}
	}

	if len(spotifyID) != 22 {
		return fmt.Errorf("invalid Spotify ID (must be 22 characters)")
	}

	// Determine output directory
	if downloadOutputDir == "" {
		downloadOutputDir = backend.GetDefaultMusicPath()
	}
	downloadOutputDir = backend.NormalizePath(downloadOutputDir)

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(downloadOutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Fetch Spotify metadata
	fmt.Printf("📍 Fetching metadata for: %s\n", spotifyID)
	spotifyURL := fmt.Sprintf("https://open.spotify.com/track/%s", spotifyID)

	trackData, err := backend.GetFilteredSpotifyData(cmd.Context(), spotifyURL, false, 0)
	if err != nil {
		return fmt.Errorf("failed to fetch metadata: %w", err)
	}

	// Extract track info
	trackInfo := extractTrackInfo(trackData)
	if trackInfo == nil {
		return fmt.Errorf("failed to extract track information")
	}

	fmt.Printf("📀 Title: %s\n", trackInfo.Title)
	fmt.Printf("🎤 Artist: %s\n", trackInfo.Artist)
	fmt.Printf("💿 Album: %s\n", trackInfo.Album)

	// Determine service
	service := downloadService
	if service == "auto" {
		service = "tidal" // Default to Tidal
	}

	// Set quality
	quality := downloadQuality
	if quality == "" {
		quality = downloadFormat
	}
	if quality == "" {
		quality = "LOSSLESS"
	}

	fmt.Printf("⬇️  Downloading from %s with quality %s...\n", service, quality)

	// Build download request
	req := DownloadRequest{
		ISRC:                 trackInfo.ISRC,
		Service:              service,
		SpotifyID:            spotifyID,
		TrackName:            trackInfo.Title,
		ArtistName:           trackInfo.Artist,
		AlbumName:            trackInfo.Album,
		AlbumArtist:          trackInfo.AlbumArtist,
		ReleaseDate:          trackInfo.ReleaseDate,
		CoverURL:             trackInfo.CoverURL,
		OutputDir:            downloadOutputDir,
		AudioFormat:          quality,
		FilenameFormat:       downloadFilenameFormat,
		TrackNumber:          downloadTrackNumber,
		UseAlbumTrackNumber:  downloadUseAlbumTrack,
		EmbedLyrics:          downloadEmbedLyrics,
		EmbedMaxQualityCover: downloadEmbedMaxQuality,
		ApiURL:               downloadTidalAPI,
	}

	// Execute download
	resp, err := downloadTrack(req)
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
}

func extractSpotifyID(url string) string {
	// Extract ID from URL like: https://open.spotify.com/track/ID
	parts := strings.Split(url, "/track/")
	if len(parts) == 2 {
		id := strings.Split(parts[1], "?")[0]
		return strings.TrimSpace(id)
	}
	return ""
}

type TrackInfo struct {
	Title       string
	Artist      string
	Album       string
	AlbumArtist string
	ReleaseDate string
	CoverURL    string
	ISRC        string
}

func extractTrackInfo(data interface{}) *TrackInfo {
	// Handle backend.TrackResponse struct
	if trackResp, ok := data.(backend.TrackResponse); ok {
		return &TrackInfo{
			Title:       trackResp.Track.Name,
			Artist:      trackResp.Track.Artists,
			Album:       trackResp.Track.AlbumName,
			AlbumArtist: trackResp.Track.AlbumArtist,
			CoverURL:    trackResp.Track.Images,
			ReleaseDate: trackResp.Track.ReleaseDate,
			ISRC:        trackResp.Track.ISRC,
		}
	}

	// Also handle map-based responses for compatibility
	if mapData, ok := data.(map[string]interface{}); ok {
		info := &TrackInfo{}

		// Extract track info from nested "track" field
		if track, ok := mapData["track"].(map[string]interface{}); ok {
			if name, ok := track["name"].(string); ok {
				info.Title = name
			}
			if isrc, ok := track["isrc"].(string); ok {
				info.ISRC = isrc
			}
			if releaseDate, ok := track["release_date"].(string); ok {
				info.ReleaseDate = releaseDate
			}

			// Extract image
			if images, ok := track["images"].([]interface{}); ok && len(images) > 0 {
				if img, ok := images[0].(map[string]interface{}); ok {
					if url, ok := img["url"].(string); ok {
						info.CoverURL = url
					}
				}
			}

			// Extract artists
			if artists, ok := track["artists"].([]interface{}); ok && len(artists) > 0 {
				if artist, ok := artists[0].(map[string]interface{}); ok {
					if name, ok := artist["name"].(string); ok {
						info.Artist = name
					}
				}
			}

			// Extract album
			if album, ok := track["album"].(map[string]interface{}); ok {
				if name, ok := album["name"].(string); ok {
					info.Album = name
				}
				if artists, ok := album["artists"].([]interface{}); ok && len(artists) > 0 {
					if artist, ok := artists[0].(map[string]interface{}); ok {
						if name, ok := artist["name"].(string); ok {
							info.AlbumArtist = name
						}
					}
				}
			}
		}

		return info
	}
	return nil
}

type DownloadRequest struct {
	ISRC                 string
	Service              string
	SpotifyID            string
	TrackName            string
	ArtistName           string
	AlbumName            string
	AlbumArtist          string
	ReleaseDate          string
	CoverURL             string
	OutputDir            string
	AudioFormat          string
	FilenameFormat       string
	TrackNumber          bool
	Position             int
	UseAlbumTrackNumber  bool
	EmbedLyrics          bool
	EmbedMaxQualityCover bool
	ApiURL               string
	SpotifyTrackNumber   int
	SpotifyDiscNumber    int
	SpotifyTotalTracks   int
	SpotifyTotalDiscs    int
	Copyright            string
	Publisher            string
}

type DownloadResponse struct {
	Success       bool
	Message       string
	File          string
	Error         string
	AlreadyExists bool
	ItemID        string
}

func downloadTrack(req DownloadRequest) (DownloadResponse, error) {
	if req.OutputDir == "" {
		req.OutputDir = "."
	} else {
		req.OutputDir = backend.NormalizePath(req.OutputDir)
	}

	if req.AudioFormat == "" {
		req.AudioFormat = "LOSSLESS"
	}

	if req.FilenameFormat == "" {
		req.FilenameFormat = "title-artist"
	}

	var filename string
	var err error

	// Check if file already exists
	if req.TrackName != "" && req.ArtistName != "" {
		expectedFilename := backend.BuildExpectedFilename(req.TrackName, req.ArtistName, req.AlbumName, req.AlbumArtist, req.ReleaseDate, req.FilenameFormat, req.TrackNumber, req.Position, req.SpotifyDiscNumber, req.UseAlbumTrackNumber)
		expectedPath := filepath.Join(req.OutputDir, expectedFilename)

		if fileInfo, err := os.Stat(expectedPath); err == nil && fileInfo.Size() > 100*1024 {
			return DownloadResponse{
				Success:       true,
				Message:       "File already exists",
				File:          expectedPath,
				AlreadyExists: true,
			}, nil
		}
	}

	// Download based on service
	switch req.Service {
	case "amazon":
		downloader := backend.NewAmazonDownloader()
		filename, err = downloader.DownloadBySpotifyID(req.SpotifyID, req.OutputDir, req.AudioFormat, req.FilenameFormat, req.TrackNumber, req.Position, req.TrackName, req.ArtistName, req.AlbumName, req.AlbumArtist, req.ReleaseDate, req.CoverURL, req.SpotifyTrackNumber, req.SpotifyDiscNumber, req.SpotifyTotalTracks, req.EmbedMaxQualityCover, req.SpotifyTotalDiscs, req.Copyright, req.Publisher, fmt.Sprintf("https://open.spotify.com/track/%s", req.SpotifyID))

	case "tidal":
		downloader := backend.NewTidalDownloader(req.ApiURL)
		filename, err = downloader.Download(req.SpotifyID, req.OutputDir, req.AudioFormat, req.FilenameFormat, req.TrackNumber, req.Position, req.TrackName, req.ArtistName, req.AlbumName, req.AlbumArtist, req.ReleaseDate, req.UseAlbumTrackNumber, req.CoverURL, req.EmbedMaxQualityCover, req.SpotifyTrackNumber, req.SpotifyDiscNumber, req.SpotifyTotalTracks, req.SpotifyTotalDiscs, req.Copyright, req.Publisher, fmt.Sprintf("https://open.spotify.com/track/%s", req.SpotifyID))

	case "qobuz":
		downloader := backend.NewQobuzDownloader()
		quality := req.AudioFormat
		if quality == "" {
			quality = "6"
		}

		// Try to get ISRC from Deezer if not provided
		if req.ISRC == "" && req.SpotifyID != "" {
			client := backend.NewSongLinkClient()
			deezerURL, err := client.GetDeezerURLFromSpotify(req.SpotifyID)
			if err == nil {
				req.ISRC, _ = backend.GetDeezerISRC(deezerURL)
			}
		}

		if req.ISRC == "" {
			return DownloadResponse{
				Success: false,
				Error:   "ISRC is required for Qobuz",
			}, fmt.Errorf("ISRC is required for Qobuz")
		}

		filename, err = downloader.DownloadByISRC(req.ISRC, req.OutputDir, quality, req.FilenameFormat, req.TrackNumber, req.Position, req.TrackName, req.ArtistName, req.AlbumName, req.AlbumArtist, req.ReleaseDate, req.UseAlbumTrackNumber, req.CoverURL, req.EmbedMaxQualityCover, req.SpotifyTrackNumber, req.SpotifyDiscNumber, req.SpotifyTotalTracks, req.SpotifyTotalDiscs, req.Copyright, req.Publisher, fmt.Sprintf("https://open.spotify.com/track/%s", req.SpotifyID))

	default:
		return DownloadResponse{
			Success: false,
			Error:   fmt.Sprintf("Unknown service: %s", req.Service),
		}, fmt.Errorf("unknown service: %s", req.Service)
	}

	if err != nil {
		// Cleanup partial file
		if filename != "" && !strings.HasPrefix(filename, "EXISTS:") {
			if _, statErr := os.Stat(filename); statErr == nil {
				os.Remove(filename)
			}
		}
		return DownloadResponse{
			Success: false,
			Error:   fmt.Sprintf("Download failed: %v", err),
		}, err
	}

	alreadyExists := false
	if strings.HasPrefix(filename, "EXISTS:") {
		alreadyExists = true
		filename = strings.TrimPrefix(filename, "EXISTS:")
	}

	// Embed lyrics if requested
	if !alreadyExists && req.SpotifyID != "" && req.EmbedLyrics && strings.HasSuffix(filename, ".flac") {
		go func(filePath, spotifyID, trackName, artistName string) {
			lyricsClient := backend.NewLyricsClient()
			lyricsResp, _, err := lyricsClient.FetchLyricsAllSources(spotifyID, trackName, artistName, 0)
			if err == nil && lyricsResp != nil && len(lyricsResp.Lines) > 0 {
				lyrics := lyricsClient.ConvertToLRC(lyricsResp, trackName, artistName)
				if lyrics != "" {
					backend.EmbedLyricsOnly(filePath, lyrics)
				}
			}
		}(filename, req.SpotifyID, req.TrackName, req.ArtistName)
	}

	message := "Download completed successfully"
	if alreadyExists {
		message = "File already exists"
	} else {
		// Add to history
		go func(fPath, track, artist, album, sID, cover string) {
			item := backend.HistoryItem{
				SpotifyID:   sID,
				Title:       track,
				Artists:     artist,
				Album:       album,
				CoverURL:    cover,
				Quality:     "Unknown",
				Format:      "FLAC",
				Path:        fPath,
				DurationStr: "--:--",
			}
			backend.AddHistoryItem(item, "SpotiFLAC")
		}(filename, req.TrackName, req.ArtistName, req.AlbumName, req.SpotifyID, req.CoverURL)
	}

	return DownloadResponse{
		Success:       true,
		Message:       message,
		File:          filename,
		AlreadyExists: alreadyExists,
	}, nil
}
