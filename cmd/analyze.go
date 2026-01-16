package cmd

import (
	"encoding/json"
	"fmt"

	"spotiflac/backend"

	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze <file-path>",
	Short: "Analyze audio file quality and metadata",
	Long: `Analyze an audio file to get detailed quality and metadata information.

Shows bitrate, sample rate, duration, and codec information.

Examples:
  spotflac analyze song.flac
  spotflac analyze song.flac --format json`,
	Args: cobra.ExactArgs(1),
	RunE: runAnalyze,
}

var (
	analyzeFormat string
)

func init() {
	analyzeCmd.Flags().StringVar(&analyzeFormat, "format", "pretty", "Output format: json or pretty")
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	fmt.Printf("🔍 Analyzing: %s\n", filePath)

	result, err := backend.AnalyzeTrack(filePath)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	if analyzeFormat == "json" {
		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to encode response: %w", err)
		}
		fmt.Println(string(jsonBytes))
	} else {
		printAnalysisResult(result)
	}

	return nil
}

func printAnalysisResult(result *backend.AnalysisResult) {
	if result == nil {
		return
	}

	fmt.Println("\n📊 Audio Analysis")
	fmt.Println("════════════════════════════════════════════")

	if result.BitDepth != "" {
		fmt.Printf("Bit Depth: %s\n", result.BitDepth)
	}

	if result.Duration > 0 {
		minutes := int(result.Duration) / 60
		seconds := int(result.Duration) % 60
		fmt.Printf("Duration: %d:%02d\n", minutes, seconds)
	}

	if result.SampleRate > 0 {
		fmt.Printf("Sample Rate: %d Hz (%.1f kHz)\n", result.SampleRate, float64(result.SampleRate)/1000)
	}

	if result.Channels > 0 {
		channelStr := "Mono"
		if result.Channels == 2 {
			channelStr = "Stereo"
		} else if result.Channels > 2 {
			channelStr = fmt.Sprintf("%d-channel", result.Channels)
		}
		fmt.Printf("Channels: %s\n", channelStr)
	}

	if result.BitsPerSample > 0 {
		fmt.Printf("Bits Per Sample: %d-bit\n", result.BitsPerSample)
	}

	if result.DynamicRange > 0 {
		fmt.Printf("Dynamic Range: %.1f dB\n", result.DynamicRange)
	}

	fmt.Printf("\n✅ Quality Rating: %s-bit/%dkHz\n",
		fmt.Sprintf("%d", result.BitsPerSample),
		result.SampleRate/1000)
}
