package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadHighScore(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "puyo-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Note: We use the real high score path for this test
	_ = filepath.Join(tempDir, "highscore.json")

	// Create a high score
	hs := &HighScore{
		Score:  10000,
		Level:  5,
		Chains: 20,
	}

	// Save using the real path (will go to ~/.puyo)
	err = SaveHighScore(hs)
	if err != nil {
		t.Errorf("Failed to save high score: %v", err)
	}

	// Load high score
	loaded, err := LoadHighScore()
	if err != nil {
		t.Fatalf("Failed to load high score: %v", err)
	}

	// Verify
	if loaded.Score != hs.Score {
		t.Errorf("Expected score %d, got %d", hs.Score, loaded.Score)
	}
	if loaded.Level != hs.Level {
		t.Errorf("Expected level %d, got %d", hs.Level, loaded.Level)
	}
	if loaded.Chains != hs.Chains {
		t.Errorf("Expected chains %d, got %d", hs.Chains, loaded.Chains)
	}

	// Clean up the actual high score file created during test
	path, _ := getHighScorePath()
	os.Remove(path)
}

func TestLoadHighScoreNotExists(t *testing.T) {
	// Try to load when file doesn't exist (should return empty high score)
	// First, remove any existing high score
	path, err := getHighScorePath()
	if err == nil {
		os.Remove(path)
	}

	hs, err := LoadHighScore()
	if err != nil {
		t.Errorf("LoadHighScore should not error when file doesn't exist: %v", err)
	}

	if hs.Score != 0 || hs.Level != 0 || hs.Chains != 0 {
		t.Error("Expected empty high score when file doesn't exist")
	}
}

func TestUpdateHighScore(t *testing.T) {
	// Clean up before test
	path, _ := getHighScorePath()
	os.Remove(path)

	game := NewGame()
	game.Score = 5000
	game.Level = 3
	game.TotalChains = 10

	// First update should be new high score
	hs, isNew, err := UpdateHighScore(game)
	if err != nil {
		t.Fatalf("UpdateHighScore failed: %v", err)
	}

	if !isNew {
		t.Error("Expected first score to be new high score")
	}

	if hs.Score != game.Score {
		t.Errorf("Expected high score %d, got %d", game.Score, hs.Score)
	}

	// Second update with lower score should not update
	game2 := NewGame()
	game2.Score = 3000
	game2.Level = 2
	game2.TotalChains = 5

	hs2, isNew2, err := UpdateHighScore(game2)
	if err != nil {
		t.Fatalf("UpdateHighScore failed: %v", err)
	}

	if isNew2 {
		t.Error("Expected lower score NOT to be new high score")
	}

	if hs2.Score != 5000 {
		t.Errorf("Expected high score to remain 5000, got %d", hs2.Score)
	}

	// Third update with higher score should update
	game3 := NewGame()
	game3.Score = 10000
	game3.Level = 8
	game3.TotalChains = 30

	hs3, isNew3, err := UpdateHighScore(game3)
	if err != nil {
		t.Fatalf("UpdateHighScore failed: %v", err)
	}

	if !isNew3 {
		t.Error("Expected higher score to be new high score")
	}

	if hs3.Score != 10000 {
		t.Errorf("Expected high score to be 10000, got %d", hs3.Score)
	}

	// Clean up
	os.Remove(path)
}

func TestGetHighScorePath(t *testing.T) {
	path, err := getHighScorePath()
	if err != nil {
		t.Fatalf("getHighScorePath failed: %v", err)
	}

	if path == "" {
		t.Error("Expected non-empty path")
	}

	// Check that path ends with expected file
	if filepath.Base(path) != "highscore.json" {
		t.Errorf("Expected path to end with 'highscore.json', got %s", filepath.Base(path))
	}

	// Check that directory is .puyo
	dir := filepath.Base(filepath.Dir(path))
	if dir != ".puyo" {
		t.Errorf("Expected directory to be '.puyo', got %s", dir)
	}
}
