package main

import (
	"github.com/gdamore/tcell/v2"
)

// Screen represents the menu screen
type Screen struct {
	screen tcell.Screen
}

// NewScreen creates a new menu screen
func NewScreen() (*Screen, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err := screen.Init(); err != nil {
		return nil, err
	}

	screen.SetStyle(tcell.StyleDefault)
	screen.Clear()

	return &Screen{
		screen: screen,
	}, nil
}

// Close closes the screen
func (s *Screen) Close() {
	s.screen.Fini()
}

// drawText draws text at the given position
func (s *Screen) drawText(x, y int, text string, style tcell.Style) {
	for i, r := range text {
		s.screen.SetContent(x+i, y, r, nil, style)
	}
}

// ShowMenu displays the color selection menu and returns the selected color count
func (s *Screen) ShowMenu() int {
	selected := 0 // 0 = 4 colors, 1 = 5 colors
	options := []string{"4色", "5色"}

	for {
		s.screen.Clear()

		titleStyle := tcell.StyleDefault.Bold(true).Foreground(tcell.ColorYellow)
		normalStyle := tcell.StyleDefault
		selectedStyle := tcell.StyleDefault.Reverse(true).Bold(true)

		// Title
		s.drawText(10, 3, "Terminal Puyo", titleStyle)

		// Menu title
		s.drawText(10, 6, "色数を選択してください:", normalStyle)

		// Options
		for i, option := range options {
			style := normalStyle
			prefix := "  "
			if i == selected {
				style = selectedStyle
				prefix = "▶ "
			}
			s.drawText(12, 8+i, prefix+option, style)
		}

		// Instructions
		instructionStyle := tcell.StyleDefault.Foreground(tcell.ColorGray)
		s.drawText(10, 12, "↑↓: 選択  Enter: 決定", instructionStyle)

		s.screen.Show()

		// Handle input
		ev := s.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyUp:
				selected = (selected - 1 + len(options)) % len(options)
			case tcell.KeyDown:
				selected = (selected + 1) % len(options)
			case tcell.KeyEnter:
				// Return color count (4 or 5)
				return selected + 4
			case tcell.KeyEscape:
				// Default to 4 colors on escape
				return 4
			}
			switch ev.Rune() {
			case 'q', 'Q':
				return 4
			}
		}
	}
}
