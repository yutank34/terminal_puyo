package main

import (
	"fmt"
	"log"
)

func main() {
	// Load high score
	highScore, err := LoadHighScore()
	if err != nil {
		log.Printf("Warning: Could not load high score: %v", err)
		highScore = &HighScore{}
	}

	// Show color selection menu
	colorCount := showColorSelectionMenu()

	// Create new game with selected color count
	game := NewGameWithColors(colorCount)
	game.HighScore = highScore

	// Create UI
	ui, err := NewUI(game)
	if err != nil {
		log.Fatalf("Failed to initialize UI: %v", err)
	}
	defer ui.Close()

	// Run game
	ui.Run()

	// Save high score if game is over
	if game.GameOver {
		newHS, isNew, err := UpdateHighScore(game)
		if err != nil {
			log.Printf("Warning: Could not save high score: %v", err)
		} else if isNew {
			fmt.Printf("\n🎉 New High Score! Score: %d, Level: %d, Chains: %d\n", newHS.Score, newHS.Level, newHS.Chains)
		} else {
			fmt.Printf("\nGame Over! Score: %d, Level: %d, Chains: %d\n", game.Score, game.Level, game.TotalChains)
			fmt.Printf("High Score: %d\n", newHS.Score)
		}
	}
}

func showColorSelectionMenu() int {
	screen, err := NewScreen()
	if err != nil {
		log.Printf("Warning: Could not initialize screen for menu: %v", err)
		return 4 // Default to 4 colors
	}
	defer screen.Close()

	return screen.ShowMenu()
}
