# 🚀 SpotiFLAC CLI - Quick Start Guide

## Installation

```bash
# Build the binary
go build -o spotflac

# Move to your PATH (optional)
sudo mv spotflac /usr/local/bin/
```

## Basic Usage

### 1. Download a Track
```bash
# From Spotify URL
spotflac download "https://open.spotify.com/track/11dFghVve3povm4FtC2oew"

# From Spotify ID
spotflac download 11dFghVve3povm4FtC2oew

# With options
spotflac download 11dFghVve3povm4FtC2oew \
  --service tidal \
  --quality LOSSLESS \
  --output ~/Music
```

### 2. Search for Music
```bash
# Search tracks (default)
spotflac search "The Beatles"

# Search albums
spotflac search "Abbey Road" --type album

# Search artists
spotflac search "Pink Floyd" --type artist

# Limit results
spotflac search "Dua Lipa" --limit 5
```

### 3. Configure Settings
```bash
# Show current configuration
spotflac config show

# Set download path
spotflac config set download-path ~/Music

# Set default service
spotflac config set downloader tidal

# Reset to defaults
spotflac config reset
```

### 4. View Download History
```bash
# List recent downloads
spotflac history list

# Search history
spotflac history search "The Beatles"

# Export as JSON
spotflac history export history.json

# Clear history
spotflac history clear
```

## Available Services

- **Tidal**: `--service tidal` (Quality: LOSSLESS, HI_RES_LOSSLESS)
- **Qobuz**: `--service qobuz` (Quality: 6, 7)
- **Amazon**: `--service amazon` (Quality: original)
- **Auto**: `--service auto` (Auto-select best available)

## Common Flags

```bash
spotflac download <ID> \
  --output ~/Music           # Output directory
  --service tidal            # Music service
  --quality LOSSLESS         # Audio quality
  --format flac              # Output format
  --folder true              # Create album folders
  --embed-lyrics true        # Embed lyrics
  --embed-max-quality-cover  # Embed high-res cover
```

## Advanced Features

### Audio Analysis
```bash
spotflac analyze ~/Music/track.flac
```

### Format Conversion
```bash
spotflac convert input.flac \
  --output-format mp3 \
  --bitrate 320
```

### Get Lyrics
```bash
spotflac lyrics 11dFghVve3povm4FtC2oew
```

### Get Album Cover
```bash
spotflac cover 11dFghVve3povm4FtC2oew
```

### Check Availability
```bash
spotflac availability 11dFghVve3povm4FtC2oew
```

## Getting Help

```bash
spotflac --help           # Show all commands
spotflac download --help  # Show download options
spotflac config --help    # Show config options
```

## Documentation

- **CLI_GUIDE.md** - Complete command reference
- **CLI_IMPLEMENTATION.md** - Architecture & design
- **PROJECT_SUMMARY.md** - Project overview
- **DEPLOYMENT_CHECKLIST.md** - Verification checklist

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Command not found | Run `go build -o spotflac` or add to PATH |
| Download fails | Check service credentials in config |
| No results | Verify Spotify track ID is correct |
| Permission denied | Run `chmod +x spotflac` |

## Version
```bash
spotflac --version
# Output: SpotiFLAC v7.0.6
```

---

**Ready to go!** Start with `spotflac --help` to explore all commands.
