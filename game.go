package main

import (
	"fmt"
	"math/rand"
	"time"
)

// Color represents the color of a puyo
type Color int

const (
	Empty Color = iota
	Red
	Green
	Blue
	Yellow
	Purple
)

// String returns the display character for a color
func (c Color) String() string {
	switch c {
	case Red:
		return "🔴"
	case Green:
		return "🟢"
	case Blue:
		return "🔵"
	case Yellow:
		return "🟡"
	case Purple:
		return "🟣"
	default:
		return "  "
	}
}

const (
	FieldWidth  = 6
	FieldHeight = 12
	MinChain    = 4 // Minimum puyos to clear
)

// Position represents a position on the field
type Position struct {
	X, Y int
}

// Puyo represents a single puyo piece
type Puyo struct {
	Color Color
}

// PuyoPair represents a falling pair of puyos
type PuyoPair struct {
	Main   Puyo
	Sub    Puyo
	Pos    Position // Position of main puyo
	Rotate int      // 0=sub on top, 1=sub on right, 2=sub on bottom, 3=sub on left
}

// Field represents the game field
type Field struct {
	Grid [FieldHeight][FieldWidth]Color
}

// NewField creates a new empty field
func NewField() *Field {
	return &Field{}
}

// IsValidPosition checks if a position is valid and empty
// Allows negative Y (above screen) for spawning puyos
func (f *Field) IsValidPosition(x, y int) bool {
	if x < 0 || x >= FieldWidth || y >= FieldHeight {
		return false
	}
	// Y < 0 is allowed (spawn area above screen)
	if y < 0 {
		return true
	}
	return f.Grid[y][x] == Empty
}

// PlacePuyo places a puyo at the given position
func (f *Field) PlacePuyo(x, y int, color Color) {
	if x >= 0 && x < FieldWidth && y >= 0 && y < FieldHeight {
		f.Grid[y][x] = color
	}
}

// GetSubPosition returns the position of the sub puyo based on rotation
func (p *PuyoPair) GetSubPosition() Position {
	switch p.Rotate {
	case 0: // Top
		return Position{p.Pos.X, p.Pos.Y - 1}
	case 1: // Right
		return Position{p.Pos.X + 1, p.Pos.Y}
	case 2: // Bottom
		return Position{p.Pos.X, p.Pos.Y + 1}
	case 3: // Left
		return Position{p.Pos.X - 1, p.Pos.Y}
	default:
		return Position{p.Pos.X, p.Pos.Y - 1}
	}
}

// GameState represents the current state of the game
type GameState int

const (
	StateNormal   GameState = iota // Normal play
	StateClearing                  // Puyos are being cleared
	StateDropping                  // Puyos are falling after clear
)

// Game represents the game state
type Game struct {
	Field           *Field
	Current         *PuyoPair
	Next            *PuyoPair
	Score           int
	Level           int
	GameOver        bool
	Paused          bool // Game is paused
	rand            *rand.Rand
	ChainCount      int
	TotalChains     int
	LinesCleared    int
	DropSpeed       time.Duration
	HighScore       *HighScore
	State           GameState
	CurrentChainNum int // Current chain number being displayed
	GroundFrames    int // Frames spent on ground (lock delay counter)
	MaxGroundFrames int // Maximum frames allowed on ground (32 for Puyo Puyo Tsu)
	ColorCount      int // Number of colors (4 or 5)
}

// TogglePause toggles the pause state
func (g *Game) TogglePause() {
	// Don't allow pause during chain animation or when game is over
	if g.State != StateNormal || g.GameOver {
		return
	}
	g.Paused = !g.Paused
}

// NewGame creates a new game
func NewGame() *Game {
	return NewGameWithColors(4) // Default to 4 colors
}

// NewGameWithColors creates a new game with specified number of colors (4 or 5)
func NewGameWithColors(colorCount int) *Game {
	if colorCount != 4 && colorCount != 5 {
		colorCount = 4 // Default to 4 if invalid
	}

	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	g := &Game{
		Field:           NewField(),
		rand:            r,
		Level:           1,
		DropSpeed:       500 * time.Millisecond,
		MaxGroundFrames: 32, // Puyo Puyo Tsu specification
		ColorCount:      colorCount,
	}

	g.Next = g.generatePuyoPair()
	g.SpawnNewPair()

	return g
}

// generatePuyoPair generates a random puyo pair
func (g *Game) generatePuyoPair() *PuyoPair {
	var colors []Color
	if g.ColorCount == 4 {
		colors = []Color{Red, Green, Blue, Yellow}
	} else {
		colors = []Color{Red, Green, Blue, Yellow, Purple}
	}
	return &PuyoPair{
		Main:   Puyo{Color: colors[g.rand.Intn(len(colors))]},
		Sub:    Puyo{Color: colors[g.rand.Intn(len(colors))]},
		Pos:    Position{X: FieldWidth / 2, Y: 0}, // Sub puyo will be at Y=-1 (above screen)
		Rotate: 0,
	}
}

// SpawnNewPair spawns a new puyo pair
func (g *Game) SpawnNewPair() {
	g.Current = g.Next
	g.Next = g.generatePuyoPair()
	g.ChainCount = 0

	// Check if spawn position is blocked (game over)
	subPos := g.Current.GetSubPosition()
	if !g.Field.IsValidPosition(g.Current.Pos.X, g.Current.Pos.Y) ||
		!g.Field.IsValidPosition(subPos.X, subPos.Y) {
		g.GameOver = true
	}
}

// CanMove checks if the current pair can move to the given position
func (g *Game) CanMove(dx, dy, rotate int) bool {
	if g.Current == nil {
		return false
	}

	newX := g.Current.Pos.X + dx
	newY := g.Current.Pos.Y + dy
	newRotate := (g.Current.Rotate + rotate + 4) % 4

	// Check main puyo
	if !g.Field.IsValidPosition(newX, newY) {
		return false
	}

	// Check sub puyo
	testPair := &PuyoPair{
		Pos:    Position{newX, newY},
		Rotate: newRotate,
	}
	subPos := testPair.GetSubPosition()

	return g.Field.IsValidPosition(subPos.X, subPos.Y)
}

// Move moves the current pair
func (g *Game) Move(dx, dy, rotate int) bool {
	// Try normal move first
	if g.CanMove(dx, dy, rotate) {
		g.Current.Pos.X += dx
		g.Current.Pos.Y += dy
		g.Current.Rotate = (g.Current.Rotate + rotate + 4) % 4

		// Reset ground timer on horizontal movement or rotation (not downward movement)
		if dx != 0 || rotate != 0 {
			g.GroundFrames = 0
		}
		return true
	}

	// If rotation failed, try wall kick (shift horizontally)
	if rotate != 0 {
		// Try shifting left
		if g.CanMove(dx-1, dy, rotate) {
			g.Current.Pos.X += dx - 1
			g.Current.Pos.Y += dy
			g.Current.Rotate = (g.Current.Rotate + rotate + 4) % 4
			g.GroundFrames = 0
			return true
		}

		// Try shifting right
		if g.CanMove(dx+1, dy, rotate) {
			g.Current.Pos.X += dx + 1
			g.Current.Pos.Y += dy
			g.Current.Rotate = (g.Current.Rotate + rotate + 4) % 4
			g.GroundFrames = 0
			return true
		}

		// Try shifting up (for floor kick)
		if g.CanMove(dx, dy-1, rotate) {
			g.Current.Pos.X += dx
			g.Current.Pos.Y += dy - 1
			g.Current.Rotate = (g.Current.Rotate + rotate + 4) % 4
			g.GroundFrames = 0
			return true
		}
	}

	return false
}

// Drop drops the current pair by one row
// Returns true if the drop was successful, false if the pair is on the ground
func (g *Game) Drop() bool {
	if g.Move(0, 1, 0) {
		// Successfully moved down, reset ground timer
		g.GroundFrames = 0
		return true
	}
	// Could not move down (on ground)
	// Note: GroundFrames is now incremented by the frame ticker in UI
	return false
}

// IsOnGround checks if the current pair is on the ground or another puyo
func (g *Game) IsOnGround() bool {
	if g.Current == nil {
		return false
	}
	// Try to move down, if it fails, we're on the ground
	return !g.CanMove(0, 1, 0)
}

// ShouldLock checks if the pair should be locked based on ground time
func (g *Game) ShouldLock() bool {
	return g.GroundFrames >= g.MaxGroundFrames
}

// HardDrop drops the pair immediately and locks it
func (g *Game) HardDrop() {
	for g.Move(0, 1, 0) {
		// Keep dropping until we hit the ground
	}
	// Immediately lock (skip grace period)
	g.GroundFrames = g.MaxGroundFrames
}

// LockPair locks the current pair to the field
func (g *Game) LockPair() {
	if g.Current == nil {
		return
	}

	// Place both puyos
	g.Field.PlacePuyo(g.Current.Pos.X, g.Current.Pos.Y, g.Current.Main.Color)

	subPos := g.Current.GetSubPosition()
	g.Field.PlacePuyo(subPos.X, subPos.Y, g.Current.Sub.Color)

	// Clear the current pair so it doesn't interfere
	g.Current = nil

	// Reset ground timer
	g.GroundFrames = 0

	// Apply gravity first
	g.State = StateDropping
	g.ChainCount = 0
	g.CurrentChainNum = 0
}

// ProcessChainStep processes one step of the chain animation
// Returns true if there are more steps to process
func (g *Game) ProcessChainStep() bool {
	switch g.State {
	case StateDropping:
		// Apply gravity
		g.applyGravity()

		// Check if there are puyos to clear
		if g.hasClearablePuyos() {
			g.State = StateClearing
			g.ChainCount++
			g.CurrentChainNum = g.ChainCount
			return true
		} else {
			// No more chains, calculate final score
			g.calculateScore()
			g.State = StateNormal
			g.SpawnNewPair()
			return false
		}

	case StateClearing:
		// Clear puyos
		cleared := g.clearPuyos()
		if cleared {
			g.TotalChains++
		}
		g.State = StateDropping
		return true

	default:
		return false
	}
}

// hasClearablePuyos checks if there are any puyos that can be cleared
func (g *Game) hasClearablePuyos() bool {
	visited := make(map[Position]bool)

	for y := 0; y < FieldHeight; y++ {
		for x := 0; x < FieldWidth; x++ {
			pos := Position{x, y}
			if g.Field.Grid[y][x] != Empty && !visited[pos] {
				group := g.findConnectedGroup(x, y, g.Field.Grid[y][x], make(map[Position]bool))

				// Mark as visited
				for p := range group {
					visited[p] = true
				}

				// Check if group is large enough to clear
				if len(group) >= MinChain {
					return true
				}
			}
		}
	}

	return false
}

// calculateScore calculates the score based on chains
func (g *Game) calculateScore() {
	if g.ChainCount > 0 {
		chainBonus := 1
		for i := 1; i < g.ChainCount; i++ {
			chainBonus *= 2
		}
		g.Score += 100 * chainBonus * g.ChainCount
		g.LinesCleared += g.ChainCount

		// Level up every 10 clears
		newLevel := (g.LinesCleared / 10) + 1
		if newLevel > g.Level {
			g.Level = newLevel
			// Increase speed with level (max speed at level 20)
			if g.Level < 20 {
				g.DropSpeed = time.Duration(float64(500*time.Millisecond) / (1.0 + float64(g.Level)*0.1))
			} else {
				g.DropSpeed = 100 * time.Millisecond
			}
		}
	}
}

// applyGravity makes puyos fall down
func (g *Game) applyGravity() {
	for x := 0; x < FieldWidth; x++ {
		writeY := FieldHeight - 1
		for y := FieldHeight - 1; y >= 0; y-- {
			if g.Field.Grid[y][x] != Empty {
				if writeY != y {
					g.Field.Grid[writeY][x] = g.Field.Grid[y][x]
					g.Field.Grid[y][x] = Empty
				}
				writeY--
			}
		}
	}
}

// clearPuyos clears connected puyos of the same color
func (g *Game) clearPuyos() bool {
	visited := make(map[Position]bool)
	cleared := false

	for y := 0; y < FieldHeight; y++ {
		for x := 0; x < FieldWidth; x++ {
			pos := Position{x, y}
			if g.Field.Grid[y][x] != Empty && !visited[pos] {
				group := g.findConnectedGroup(x, y, g.Field.Grid[y][x], make(map[Position]bool))

				// Mark as visited
				for p := range group {
					visited[p] = true
				}

				// Clear if group is large enough
				if len(group) >= MinChain {
					for p := range group {
						g.Field.Grid[p.Y][p.X] = Empty
					}
					cleared = true
				}
			}
		}
	}

	return cleared
}

// findConnectedGroup finds all connected puyos of the same color
func (g *Game) findConnectedGroup(x, y int, color Color, visited map[Position]bool) map[Position]bool {
	pos := Position{x, y}

	if x < 0 || x >= FieldWidth || y < 0 || y >= FieldHeight {
		return visited
	}

	if g.Field.Grid[y][x] != color || visited[pos] {
		return visited
	}

	visited[pos] = true

	// Check all 4 directions
	g.findConnectedGroup(x+1, y, color, visited)
	g.findConnectedGroup(x-1, y, color, visited)
	g.findConnectedGroup(x, y+1, color, visited)
	g.findConnectedGroup(x, y-1, color, visited)

	return visited
}

// Display prints the current game state (for debugging)
func (g *Game) Display() {
	fmt.Println("Score:", g.Score, "Level:", g.Level, "Chains:", g.TotalChains)

	// Create a copy of the field to overlay the current pair
	display := g.Field.Grid

	if g.Current != nil {
		subPos := g.Current.GetSubPosition()
		if subPos.Y >= 0 && subPos.Y < FieldHeight && subPos.X >= 0 && subPos.X < FieldWidth {
			display[subPos.Y][subPos.X] = g.Current.Sub.Color
		}
		if g.Current.Pos.Y >= 0 && g.Current.Pos.Y < FieldHeight {
			display[g.Current.Pos.Y][g.Current.Pos.X] = g.Current.Main.Color
		}
	}

	// Print field
	for y := 0; y < FieldHeight; y++ {
		fmt.Print("|")
		for x := 0; x < FieldWidth; x++ {
			fmt.Print(Color(display[y][x]).String())
		}
		fmt.Println("|")
	}

	// Print bottom border
	for i := 0; i < FieldWidth*2+2; i++ {
		fmt.Print("=")
	}
	fmt.Println()

	// Print next pair
	if g.Next != nil {
		fmt.Println("Next:", g.Next.Main.Color.String(), g.Next.Sub.Color.String())
	}
}
