package main

import (
	"os"

	"spotiflac/backend"
	"spotiflac/cmd"
)

func main() {
	// Initialize history database
	if err := backend.InitHistoryDB("SpotiFLAC"); err != nil {
		// Non-fatal initialization error
	}
	defer backend.CloseHistoryDB()

	// Execute root command
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
