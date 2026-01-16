package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"spotiflac/backend"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage SpotiFLAC configuration",
	Long: `View and modify SpotiFLAC settings and configuration.

Configuration is stored in the platform-specific config directory.

Examples:
  spotflac config show
  spotflac config set download-path ~/Music
  spotflac config set downloader tidal
  spotflac config get download-path`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := getConfigPath()
		config, err := loadConfig(configPath)
		if err != nil {
			fmt.Println("📋 Default configuration (no config file found):")
		} else {
			fmt.Println("📋 Current configuration:")
		}

		defaultConfig := getDefaultConfig()
		if config == nil {
			config = defaultConfig
		} else {
			// Merge with defaults
			for k, v := range defaultConfig {
				if _, exists := config[k]; !exists {
					config[k] = v
				}
			}
		}

		fmt.Println("────────────────────────────────────────────")
		for key, value := range config {
			fmt.Printf("%s: %v\n", key, value)
		}

		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		configPath := getConfigPath()
		config, _ := loadConfig(configPath)

		if config == nil {
			config = getDefaultConfig()
		}

		if value, exists := config[key]; exists {
			fmt.Printf("%s = %v\n", key, value)
			return nil
		}

		// Check defaults
		defaults := getDefaultConfig()
		if value, exists := defaults[key]; exists {
			fmt.Printf("%s = %v (default)\n", key, value)
			return nil
		}

		return fmt.Errorf("unknown configuration key: %s", key)
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		configPath := getConfigPath()
		config, _ := loadConfig(configPath)
		if config == nil {
			config = make(map[string]interface{})
		}

		// Validate and set value
		switch key {
		case "download-path":
			// Expand home directory
			expanded, err := expandPath(value)
			if err != nil {
				return fmt.Errorf("invalid path: %w", err)
			}
			if err := os.MkdirAll(expanded, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			config[key] = expanded
			fmt.Printf("✅ download-path set to: %s\n", expanded)

		case "downloader":
			if !isValidDownloader(value) {
				return fmt.Errorf("invalid downloader: %s (must be: auto, tidal, qobuz, amazon)", value)
			}
			config[key] = value
			fmt.Printf("✅ downloader set to: %s\n", value)

		case "tidal-quality":
			if !isValidTidalQuality(value) {
				return fmt.Errorf("invalid tidal-quality: %s (must be: LOSSLESS or HI_RES_LOSSLESS)", value)
			}
			config[key] = value
			fmt.Printf("✅ tidal-quality set to: %s\n", value)

		case "qobuz-quality":
			if !isValidQobuzQuality(value) {
				return fmt.Errorf("invalid qobuz-quality: %s (must be: 6 or 7)", value)
			}
			config[key] = value
			fmt.Printf("✅ qobuz-quality set to: %s\n", value)

		case "filename-format":
			if !isValidFilenameFormat(value) {
				return fmt.Errorf("invalid filename-format: %s", value)
			}
			config[key] = value
			fmt.Printf("✅ filename-format set to: %s\n", value)

		case "folder-structure":
			if !isValidFolderStructure(value) {
				return fmt.Errorf("invalid folder-structure: %s", value)
			}
			config[key] = value
			fmt.Printf("✅ folder-structure set to: %s\n", value)

		case "embed-lyrics":
			config[key] = value == "true" || value == "yes" || value == "1"
			fmt.Printf("✅ embed-lyrics set to: %v\n", config[key])

		case "track-number":
			config[key] = value == "true" || value == "yes" || value == "1"
			fmt.Printf("✅ track-number set to: %v\n", config[key])

		default:
			return fmt.Errorf("unknown configuration key: %s", key)
		}

		// Save config
		if err := saveConfig(configPath, config); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		return nil
	},
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration to defaults",
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := getConfigPath()
		if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete config file: %w", err)
		}
		fmt.Println("✅ Configuration reset to defaults")
		return nil
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration file path",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(getConfigPath())
		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configResetCmd)
	configCmd.AddCommand(configPathCmd)
}

func getConfigPath() string {
	dir, err := backend.GetFFmpegDir()
	if err != nil {
		// Fallback to home directory
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".spotiflac")
		os.MkdirAll(dir, 0755)
	}
	return filepath.Join(dir, "config.json")
}

func loadConfig(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config, nil
}

func saveConfig(path string, config map[string]interface{}) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func getDefaultConfig() map[string]interface{} {
	defaultPath := backend.GetDefaultMusicPath()
	return map[string]interface{}{
		"download-path":     defaultPath,
		"downloader":        "auto",
		"tidal-quality":     "LOSSLESS",
		"qobuz-quality":     "6",
		"filename-format":   "title-artist",
		"folder-structure":  "none",
		"embed-lyrics":      false,
		"embed-max-quality": false,
		"track-number":      false,
	}
}

func expandPath(path string) (string, error) {
	if path == "~" || path == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return home, nil
	}
	return os.ExpandEnv(path), nil
}

func isValidDownloader(value string) bool {
	valid := map[string]bool{
		"auto":   true,
		"tidal":  true,
		"qobuz":  true,
		"amazon": true,
	}
	return valid[value]
}

func isValidTidalQuality(value string) bool {
	valid := map[string]bool{
		"LOSSLESS":        true,
		"HI_RES_LOSSLESS": true,
	}
	return valid[value]
}

func isValidQobuzQuality(value string) bool {
	valid := map[string]bool{
		"6": true,
		"7": true,
	}
	return valid[value]
}

func isValidFilenameFormat(value string) bool {
	valid := map[string]bool{
		"title":              true,
		"title-artist":       true,
		"artist-title":       true,
		"track-title":        true,
		"artist-album-title": true,
	}
	return valid[value]
}

func isValidFolderStructure(value string) bool {
	valid := map[string]bool{
		"none":                    true,
		"artist":                  true,
		"album":                   true,
		"artist-album":            true,
		"year-album":              true,
		"year-artist-album":       true,
		"album-artist":            true,
		"album-artist-album":      true,
		"album-artist-year-album": true,
		"year":                    true,
		"year-artist":             true,
	}
	return valid[value]
}
