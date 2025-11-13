package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// HighScore represents a high score record
type HighScore struct {
	Score  int `json:"score"`
	Level  int `json:"level"`
	Chains int `json:"chains"`
}

// getHighScorePath returns the path to the high score file
func getHighScorePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(homeDir, ".puyo")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(configDir, "highscore.json"), nil
}

// LoadHighScore loads the high score from disk
func LoadHighScore() (*HighScore, error) {
	path, err := getHighScorePath()
	if err != nil {
		return &HighScore{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &HighScore{}, nil
		}
		return nil, err
	}

	var hs HighScore
	if err := json.Unmarshal(data, &hs); err != nil {
		return &HighScore{}, nil
	}

	return &hs, nil
}

// SaveHighScore saves the high score to disk
func SaveHighScore(hs *HighScore) error {
	path, err := getHighScorePath()
	if err != nil {
		return err
	}

	data, err := json.Marshal(hs)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// UpdateHighScore updates the high score if the current score is higher
func UpdateHighScore(game *Game) (*HighScore, bool, error) {
	current, err := LoadHighScore()
	if err != nil {
		return nil, false, err
	}

	isNew := game.Score > current.Score

	if isNew {
		newHS := &HighScore{
			Score:  game.Score,
			Level:  game.Level,
			Chains: game.TotalChains,
		}

		if err := SaveHighScore(newHS); err != nil {
			return nil, false, err
		}

		return newHS, true, nil
	}

	return current, false, nil
}
