package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"time"
)

// UI represents the terminal UI
type UI struct {
	screen tcell.Screen
	game   *Game
}

// NewUI creates a new UI
func NewUI(game *Game) (*UI, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err := screen.Init(); err != nil {
		return nil, err
	}

	screen.SetStyle(tcell.StyleDefault)
	screen.Clear()

	return &UI{
		screen: screen,
		game:   game,
	}, nil
}

// Close closes the UI
func (ui *UI) Close() {
	ui.screen.Fini()
}

// drawText draws text at the given position
func (ui *UI) drawText(x, y int, text string, style tcell.Style) {
	for i, r := range text {
		ui.screen.SetContent(x+i, y, r, nil, style)
	}
}

// Draw draws the game state
func (ui *UI) Draw() {
	ui.screen.Clear()

	style := tcell.StyleDefault
	titleStyle := tcell.StyleDefault.Bold(true)
	headerStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)

	// Title
	ui.drawText(2, 1, "Terminal Puyo", titleStyle)

	// Score and stats
	ui.drawText(2, 3, fmt.Sprintf("Score: %d", ui.game.Score), headerStyle)
	ui.drawText(2, 4, fmt.Sprintf("Level: %d", ui.game.Level), headerStyle)
	ui.drawText(2, 5, fmt.Sprintf("Chains: %d", ui.game.TotalChains), headerStyle)
	ui.drawText(2, 6, fmt.Sprintf("Colors: %d", ui.game.ColorCount), headerStyle)

	// High score
	if ui.game.HighScore != nil && ui.game.HighScore.Score > 0 {
		hsStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)
		ui.drawText(2, 7, fmt.Sprintf("High Score: %d", ui.game.HighScore.Score), hsStyle)
	}

	// Create a copy of the field to overlay the current pair
	display := ui.game.Field.Grid

	if ui.game.Current != nil {
		subPos := ui.game.Current.GetSubPosition()
		if subPos.Y >= 0 && subPos.Y < FieldHeight && subPos.X >= 0 && subPos.X < FieldWidth {
			display[subPos.Y][subPos.X] = ui.game.Current.Sub.Color
		}
		if ui.game.Current.Pos.Y >= 0 && ui.game.Current.Pos.Y < FieldHeight {
			display[ui.game.Current.Pos.Y][ui.game.Current.Pos.X] = ui.game.Current.Main.Color
		}
	}

	// Draw field border and content
	startY := 8
	startX := 2

	// Top border
	ui.drawText(startX, startY, "┌", style)
	for i := 0; i < FieldWidth*2; i++ {
		ui.drawText(startX+1+i, startY, "─", style)
	}
	ui.drawText(startX+FieldWidth*2+1, startY, "┐", style)

	// Field content
	for y := 0; y < FieldHeight; y++ {
		ui.drawText(startX, startY+1+y, "│", style)
		for x := 0; x < FieldWidth; x++ {
			color := display[y][x]
			cellStyle := style
			char := "  "

			switch color {
			case Red:
				cellStyle = style.Foreground(tcell.ColorRed)
				char = "●"
			case Green:
				cellStyle = style.Foreground(tcell.ColorGreen)
				char = "●"
			case Blue:
				cellStyle = style.Foreground(tcell.ColorBlue)
				char = "●"
			case Yellow:
				cellStyle = style.Foreground(tcell.ColorYellow)
				char = "●"
			case Purple:
				cellStyle = style.Foreground(tcell.ColorPurple)
				char = "●"
			}

			ui.drawText(startX+1+x*2, startY+1+y, char+" ", cellStyle)
		}
		ui.drawText(startX+FieldWidth*2+1, startY+1+y, "│", style)
	}

	// Bottom border
	ui.drawText(startX, startY+FieldHeight+1, "└", style)
	for i := 0; i < FieldWidth*2; i++ {
		ui.drawText(startX+1+i, startY+FieldHeight+1, "─", style)
	}
	ui.drawText(startX+FieldWidth*2+1, startY+FieldHeight+1, "┘", style)

	// Next puyo
	nextY := startY + 2
	nextX := startX + FieldWidth*2 + 5

	ui.drawText(nextX, nextY, "Next:", headerStyle)
	if ui.game.Next != nil {
		nextStyle := style.Foreground(getColorForPuyo(ui.game.Next.Main.Color))
		ui.drawText(nextX, nextY+1, "●", nextStyle)

		nextStyle = style.Foreground(getColorForPuyo(ui.game.Next.Sub.Color))
		ui.drawText(nextX, nextY+2, "●", nextStyle)
	}

	// Chain display
	if ui.game.State != StateNormal && ui.game.CurrentChainNum > 0 {
		chainY := startY + FieldHeight/2 - 2
		chainX := startX + FieldWidth - 2
		chainStyle := style.Foreground(tcell.ColorYellow).Bold(true)
		chainText := fmt.Sprintf("%d CHAIN!", ui.game.CurrentChainNum)
		ui.drawText(chainX, chainY, chainText, chainStyle)
	}

	// Controls
	controlsY := startY + 6
	ui.drawText(nextX, controlsY, "Controls:", headerStyle)
	ui.drawText(nextX, controlsY+1, "←→: Move", style)
	ui.drawText(nextX, controlsY+2, "↓: Drop", style)
	ui.drawText(nextX, controlsY+3, "Z/X: Rotate", style)
	ui.drawText(nextX, controlsY+4, "P: Pause", style)
	ui.drawText(nextX, controlsY+5, "Q: Quit", style)

	// Pause message
	if ui.game.Paused {
		msgY := startY + FieldHeight/2
		msgX := startX + 2
		pauseStyle := style.Foreground(tcell.ColorAqua).Bold(true)
		ui.drawText(msgX, msgY, "PAUSED", pauseStyle)
		ui.drawText(msgX-2, msgY+2, "Press P to resume", style)
	}

	// Game over message
	if ui.game.GameOver {
		msgY := startY + FieldHeight/2
		msgX := startX + 3
		gameOverStyle := style.Foreground(tcell.ColorRed).Bold(true)
		ui.drawText(msgX, msgY, "GAME OVER!", gameOverStyle)
		ui.drawText(msgX-2, msgY+2, "Press R to restart", style)
		ui.drawText(msgX-2, msgY+3, "Press Q to quit", style)
	}

	ui.screen.Show()
}

// getColorForPuyo returns the tcell color for a puyo color
func getColorForPuyo(c Color) tcell.Color {
	switch c {
	case Red:
		return tcell.ColorRed
	case Green:
		return tcell.ColorGreen
	case Blue:
		return tcell.ColorBlue
	case Yellow:
		return tcell.ColorYellow
	case Purple:
		return tcell.ColorPurple
	default:
		return tcell.ColorWhite
	}
}

// Run runs the game loop
func (ui *UI) Run() {
	ticker := time.NewTicker(ui.game.DropSpeed)
	defer ticker.Stop()

	// Frame ticker for ground time counting (1/60 second = ~16.67ms)
	frameTicker := time.NewTicker(time.Second / 60)
	defer frameTicker.Stop()

	// Chain animation ticker (faster than normal drop)
	chainTicker := time.NewTicker(300 * time.Millisecond)
	defer chainTicker.Stop()

	// Input channel
	eventChan := make(chan tcell.Event)
	go func() {
		for {
			eventChan <- ui.screen.PollEvent()
		}
	}()

	ui.Draw()

	for {
		select {
		case <-ticker.C:
			if !ui.game.GameOver && !ui.game.Paused && ui.game.State == StateNormal {
				// Update ticker speed if level changed
				ticker.Reset(ui.game.DropSpeed)

				// Try to drop
				ui.game.Drop()

				// Check if we should lock the pair
				if ui.game.ShouldLock() {
					ui.game.LockPair()
				}

				ui.Draw()
			}

		case <-frameTicker.C:
			if !ui.game.GameOver && !ui.game.Paused && ui.game.State == StateNormal {
				// Count ground frames at 60fps
				if ui.game.IsOnGround() {
					ui.game.GroundFrames++

					// Check if we should lock the pair
					if ui.game.ShouldLock() {
						ui.game.LockPair()
						ui.Draw()
					}
				}
			}

		case <-chainTicker.C:
			if !ui.game.GameOver && !ui.game.Paused && ui.game.State != StateNormal {
				// Process chain animation step by step
				hasMore := ui.game.ProcessChainStep()
				ui.Draw()

				if !hasMore {
					// Chain finished, reset to normal ticker
					ticker.Reset(ui.game.DropSpeed)
				}
			}

		case ev := <-eventChan:
			switch ev := ev.(type) {
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyEscape || ev.Rune() == 'q' || ev.Rune() == 'Q' {
					return
				}

				// Handle pause
				if ev.Rune() == 'p' || ev.Rune() == 'P' {
					ui.game.TogglePause()
					ui.Draw()
					continue
				}

				if ui.game.GameOver {
					if ev.Rune() == 'r' || ev.Rune() == 'R' {
						// Save the high score and color count before restarting
						oldHighScore := ui.game.HighScore
						oldColorCount := ui.game.ColorCount
						ui.game = NewGameWithColors(oldColorCount)
						ui.game.HighScore = oldHighScore
						ui.Draw()
					}
					continue
				}

				// Ignore input during chain animation or when paused
				if ui.game.State != StateNormal || ui.game.Paused {
					continue
				}

				switch ev.Key() {
				case tcell.KeyLeft:
					ui.game.Move(-1, 0, 0)
					ui.Draw()
				case tcell.KeyRight:
					ui.game.Move(1, 0, 0)
					ui.Draw()
				case tcell.KeyDown:
					// Soft drop - drop quickly (relies on key repeat)
					ui.game.Drop()

					// Check if we should lock the pair
					if ui.game.ShouldLock() {
						ui.game.LockPair()
					}
					ui.Draw()
				case tcell.KeyRune:
					switch ev.Rune() {
					case 'z', 'Z':
						ui.game.Move(0, 0, -1) // Rotate counter-clockwise
						ui.Draw()
					case 'x', 'X':
						ui.game.Move(0, 0, 1) // Rotate clockwise
						ui.Draw()
					}
				}

			case *tcell.EventResize:
				ui.screen.Sync()
				ui.Draw()
			}
		}
	}
}
