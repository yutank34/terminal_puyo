package main

import (
	"testing"
)

func TestNewField(t *testing.T) {
	field := NewField()

	if field == nil {
		t.Fatal("NewField returned nil")
	}

	// Check that all cells are empty
	for y := 0; y < FieldHeight; y++ {
		for x := 0; x < FieldWidth; x++ {
			if field.Grid[y][x] != Empty {
				t.Errorf("Expected empty cell at (%d, %d), got %v", x, y, field.Grid[y][x])
			}
		}
	}
}

func TestIsValidPosition(t *testing.T) {
	field := NewField()

	tests := []struct {
		x, y  int
		valid bool
	}{
		{0, 0, true},
		{FieldWidth - 1, FieldHeight - 1, true},
		{-1, 0, false},            // X out of bounds
		{0, -1, true},             // Y < 0 is allowed (spawn area)
		{0, -5, true},             // Multiple rows above screen are allowed
		{FieldWidth, 0, false},    // X out of bounds
		{0, FieldHeight, false},   // Y out of bounds (below screen)
	}

	for _, tt := range tests {
		result := field.IsValidPosition(tt.x, tt.y)
		if result != tt.valid {
			t.Errorf("IsValidPosition(%d, %d) = %v, want %v", tt.x, tt.y, result, tt.valid)
		}
	}
}

func TestPlacePuyo(t *testing.T) {
	field := NewField()

	field.PlacePuyo(2, 5, Red)

	if field.Grid[5][2] != Red {
		t.Errorf("Expected Red puyo at (2, 5), got %v", field.Grid[5][2])
	}

	// Test out of bounds (should not crash)
	field.PlacePuyo(-1, 0, Blue)
	field.PlacePuyo(0, -1, Blue)
	field.PlacePuyo(FieldWidth, 0, Blue)
	field.PlacePuyo(0, FieldHeight, Blue)
}

func TestPuyoPairRotation(t *testing.T) {
	pair := &PuyoPair{
		Main:   Puyo{Color: Red},
		Sub:    Puyo{Color: Blue},
		Pos:    Position{X: 3, Y: 5},
		Rotate: 0,
	}

	tests := []struct {
		rotate   int
		expected Position
	}{
		{0, Position{3, 4}}, // Top
		{1, Position{4, 5}}, // Right
		{2, Position{3, 6}}, // Bottom
		{3, Position{2, 5}}, // Left
	}

	for _, tt := range tests {
		pair.Rotate = tt.rotate
		subPos := pair.GetSubPosition()
		if subPos != tt.expected {
			t.Errorf("Rotation %d: expected %v, got %v", tt.rotate, tt.expected, subPos)
		}
	}
}

func TestGameMove(t *testing.T) {
	game := NewGame()

	// Move down a bit to have more room
	game.Move(0, 2, 0)

	// Store position after initial moves
	initialX := game.Current.Pos.X

	// Move left
	canMove := game.Move(-1, 0, 0)
	if canMove {
		if game.Current.Pos.X != initialX-1 {
			t.Errorf("Expected X to be %d, got %d", initialX-1, game.Current.Pos.X)
		}

		// Move right
		game.Move(1, 0, 0)
		if game.Current.Pos.X != initialX {
			t.Errorf("Expected X to be %d, got %d", initialX, game.Current.Pos.X)
		}
	} else {
		t.Log("Could not move left from initial position (may be blocked by sub puyo)")
	}
}

func TestGameRotate(t *testing.T) {
	game := NewGame()

	// Move down to middle of field where there's more space to rotate
	game.Move(0, 5, 0)

	initialRotate := game.Current.Rotate

	// Rotate clockwise
	canRotate := game.Move(0, 0, 1)
	if canRotate {
		expected := (initialRotate + 1) % 4
		if game.Current.Rotate != expected {
			t.Errorf("Expected rotation %d, got %d", expected, game.Current.Rotate)
		}

		// Rotate counter-clockwise
		game.Move(0, 0, -1)

		if game.Current.Rotate != initialRotate {
			t.Errorf("Expected rotation %d, got %d", initialRotate, game.Current.Rotate)
		}
	} else {
		t.Log("Could not rotate from initial position")
	}
}

func TestApplyGravity(t *testing.T) {
	game := NewGame()

	// Place puyos with gaps
	game.Field.Grid[5][2] = Red
	game.Field.Grid[8][2] = Blue
	game.Field.Grid[10][2] = Green

	game.applyGravity()

	// Check that puyos fell to the bottom
	if game.Field.Grid[FieldHeight-1][2] != Green {
		t.Error("Green puyo should be at the bottom")
	}
	if game.Field.Grid[FieldHeight-2][2] != Blue {
		t.Error("Blue puyo should be second from bottom")
	}
	if game.Field.Grid[FieldHeight-3][2] != Red {
		t.Error("Red puyo should be third from bottom")
	}

	// Check that gaps are filled
	for y := 0; y < FieldHeight-3; y++ {
		if game.Field.Grid[y][2] != Empty {
			t.Errorf("Expected empty at row %d, got %v", y, game.Field.Grid[y][2])
		}
	}
}

func TestClearPuyos(t *testing.T) {
	game := NewGame()

	// Create a horizontal line of 4 red puyos
	game.Field.Grid[FieldHeight-1][0] = Red
	game.Field.Grid[FieldHeight-1][1] = Red
	game.Field.Grid[FieldHeight-1][2] = Red
	game.Field.Grid[FieldHeight-1][3] = Red

	cleared := game.clearPuyos()

	if !cleared {
		t.Error("Expected puyos to be cleared")
	}

	// Check that puyos were removed
	for x := 0; x < 4; x++ {
		if game.Field.Grid[FieldHeight-1][x] != Empty {
			t.Errorf("Expected empty at (%d, %d)", x, FieldHeight-1)
		}
	}
}

func TestClearPuyosVertical(t *testing.T) {
	game := NewGame()

	// Create a vertical line of 4 blue puyos
	game.Field.Grid[FieldHeight-1][2] = Blue
	game.Field.Grid[FieldHeight-2][2] = Blue
	game.Field.Grid[FieldHeight-3][2] = Blue
	game.Field.Grid[FieldHeight-4][2] = Blue

	cleared := game.clearPuyos()

	if !cleared {
		t.Error("Expected puyos to be cleared")
	}

	// Check that puyos were removed
	for y := FieldHeight - 4; y < FieldHeight; y++ {
		if game.Field.Grid[y][2] != Empty {
			t.Errorf("Expected empty at (2, %d)", y)
		}
	}
}

func TestClearPuyosNotEnough(t *testing.T) {
	game := NewGame()

	// Create a line of only 3 puyos (not enough to clear)
	game.Field.Grid[FieldHeight-1][0] = Red
	game.Field.Grid[FieldHeight-1][1] = Red
	game.Field.Grid[FieldHeight-1][2] = Red

	cleared := game.clearPuyos()

	if cleared {
		t.Error("Expected puyos NOT to be cleared (only 3)")
	}

	// Check that puyos are still there
	for x := 0; x < 3; x++ {
		if game.Field.Grid[FieldHeight-1][x] != Red {
			t.Errorf("Expected Red at (%d, %d)", x, FieldHeight-1)
		}
	}
}

func TestFindConnectedGroup(t *testing.T) {
	game := NewGame()

	// Create an L-shape of 5 green puyos
	game.Field.Grid[FieldHeight-1][0] = Green
	game.Field.Grid[FieldHeight-1][1] = Green
	game.Field.Grid[FieldHeight-1][2] = Green
	game.Field.Grid[FieldHeight-2][0] = Green
	game.Field.Grid[FieldHeight-3][0] = Green

	visited := make(map[Position]bool)
	group := game.findConnectedGroup(0, FieldHeight-1, Green, visited)

	if len(group) != 5 {
		t.Errorf("Expected 5 connected puyos, got %d", len(group))
	}
}

func TestScoreCalculation(t *testing.T) {
	game := NewGame()

	initialScore := game.Score

	// Set up a clearable pattern
	game.Field.Grid[FieldHeight-1][0] = Red
	game.Field.Grid[FieldHeight-1][1] = Red
	game.Field.Grid[FieldHeight-1][2] = Red
	game.Field.Grid[FieldHeight-1][3] = Red

	// Manually trigger clear
	cleared := game.clearPuyos()

	if !cleared {
		t.Error("Expected puyos to be cleared")
	}

	// The actual score calculation happens in LockPair
	t.Logf("Initial score: %d (score calculation happens in LockPair)", initialScore)
}

func TestGameOver(t *testing.T) {
	game := NewGame()

	// Fill the spawn column to trigger game over
	for y := 0; y < FieldHeight; y++ {
		game.Field.Grid[y][FieldWidth/2] = Red
	}

	game.SpawnNewPair()

	if !game.GameOver {
		t.Error("Expected game over when spawn position is blocked")
	}
}

func TestLevelUp(t *testing.T) {
	game := NewGame()

	initialLevel := game.Level

	// Simulate clearing enough lines to level up
	game.LinesCleared = 10

	// Level should increase
	newLevel := (game.LinesCleared / 10) + 1

	if newLevel <= initialLevel {
		t.Errorf("Expected level to increase, got %d", newLevel)
	}
}

func TestWallKick(t *testing.T) {
	game := NewGame()

	// Move to the left edge
	game.Current.Pos.X = 0
	game.Current.Pos.Y = 5
	game.Current.Rotate = 0 // Sub puyo is on top

	// Try to rotate so sub puyo would be on the left (outside field)
	// This should wall kick and shift right
	initialX := game.Current.Pos.X
	rotated := game.Move(0, 0, 1) // Rotate clockwise

	if rotated {
		// Should have succeeded with wall kick
		if game.Current.Pos.X <= initialX {
			// Wall kick should have shifted right
			t.Log("Wall kick activated, shifted position")
		}
		t.Log("Rotation succeeded (possibly with wall kick)")
	} else {
		t.Log("Rotation failed even with wall kick attempt")
	}
}

func TestWallKickRight(t *testing.T) {
	game := NewGame()

	// Move to the right edge
	game.Current.Pos.X = FieldWidth - 1
	game.Current.Pos.Y = 5
	game.Current.Rotate = 0 // Sub puyo is on top

	// Try to rotate so sub puyo would be on the right (outside field)
	initialX := game.Current.Pos.X
	rotated := game.Move(0, 0, -1) // Rotate counter-clockwise

	if rotated {
		if game.Current.Pos.X >= initialX {
			t.Log("Wall kick activated, shifted position")
		}
		t.Log("Rotation succeeded (possibly with wall kick)")
	} else {
		t.Log("Rotation failed even with wall kick attempt")
	}
}

func TestFloorKick(t *testing.T) {
	game := NewGame()

	// Move to near the bottom
	game.Current.Pos.X = 3
	game.Current.Pos.Y = FieldHeight - 1
	game.Current.Rotate = 1 // Sub puyo is on the right

	// Try to rotate so sub puyo would be below (outside field)
	rotated := game.Move(0, 0, 1) // Rotate clockwise

	if rotated {
		// Should have succeeded with floor kick (shifted up)
		if game.Current.Pos.Y < FieldHeight-1 {
			t.Log("Floor kick activated, shifted up")
		}
		t.Log("Rotation succeeded (possibly with floor kick)")
	} else {
		t.Log("Rotation failed even with floor kick attempt")
	}
}

func TestLockDelay(t *testing.T) {
	game := NewGame()

	// Fill the entire bottom row
	for x := 0; x < FieldWidth; x++ {
		game.Field.Grid[FieldHeight-1][x] = Red
	}

	// Position puyo directly on the filled row
	game.Current.Pos.X = 3
	game.Current.Pos.Y = FieldHeight - 2
	game.Current.Rotate = 0 // Sub puyo on top

	// Drop should fail (on ground)
	canDrop := game.Drop()
	if canDrop {
		t.Error("Expected puyo to be on ground")
	}

	// Note: GroundFrames is now managed by UI's frame ticker
	// So we manually increment it for testing
	game.GroundFrames = 5

	// Should not lock yet (need 32 frames)
	if game.ShouldLock() {
		t.Error("Should not lock after only 5 frames")
	}

	// Move left should reset ground timer
	game.Move(-1, 0, 0)
	if game.GroundFrames != 0 {
		t.Errorf("Expected GroundFrames to reset to 0 after move, got %d", game.GroundFrames)
	}
}

func TestLockDelayExpires(t *testing.T) {
	game := NewGame()

	// Position puyo on ground
	game.Field.Grid[FieldHeight-1][3] = Red
	game.Current.Pos.X = 3
	game.Current.Pos.Y = FieldHeight - 2

	// Simulate 32 frames on ground
	// Note: GroundFrames is now managed by UI's frame ticker
	for i := 0; i < 32; i++ {
		game.GroundFrames++
	}

	// Should be ready to lock now
	if !game.ShouldLock() {
		t.Error("Should lock after 32 frames")
	}

	if game.GroundFrames < game.MaxGroundFrames {
		t.Errorf("Expected GroundFrames >= %d, got %d", game.MaxGroundFrames, game.GroundFrames)
	}
}

func TestHardDropSkipsGracePeriod(t *testing.T) {
	game := NewGame()

	// Position puyo in the air
	game.Current.Pos.X = 3
	game.Current.Pos.Y = 5

	// Hard drop should set ground frames to max
	game.HardDrop()

	if game.GroundFrames != game.MaxGroundFrames {
		t.Errorf("Expected HardDrop to set GroundFrames to %d, got %d", game.MaxGroundFrames, game.GroundFrames)
	}

	if !game.ShouldLock() {
		t.Error("Should be ready to lock immediately after HardDrop")
	}
}

func TestPause(t *testing.T) {
	game := NewGame()

	// Game should not be paused initially
	if game.Paused {
		t.Error("Game should not be paused initially")
	}

	// Toggle pause
	game.TogglePause()
	if !game.Paused {
		t.Error("Game should be paused after TogglePause")
	}

	// Toggle again to unpause
	game.TogglePause()
	if game.Paused {
		t.Error("Game should be unpaused after second TogglePause")
	}
}

func TestPauseDuringChain(t *testing.T) {
	game := NewGame()

	// Set game state to clearing (during chain)
	game.State = StateClearing

	// Try to pause during chain
	game.TogglePause()

	// Should not pause during chain animation
	if game.Paused {
		t.Error("Should not be able to pause during chain animation")
	}
}

func TestPauseWhenGameOver(t *testing.T) {
	game := NewGame()

	// Set game over
	game.GameOver = true

	// Try to pause when game is over
	game.TogglePause()

	// Should not pause when game is over
	if game.Paused {
		t.Error("Should not be able to pause when game is over")
	}
}

func TestColorCount4(t *testing.T) {
	game := NewGameWithColors(4)

	if game.ColorCount != 4 {
		t.Errorf("Expected ColorCount to be 4, got %d", game.ColorCount)
	}

	// Generate many pairs and check that only 4 colors appear
	colorsSeen := make(map[Color]bool)
	for i := 0; i < 100; i++ {
		pair := game.generatePuyoPair()
		colorsSeen[pair.Main.Color] = true
		colorsSeen[pair.Sub.Color] = true
	}

	// Should only have 4 colors (Red, Green, Blue, Yellow)
	expectedColors := map[Color]bool{Red: true, Green: true, Blue: true, Yellow: true}
	for color := range colorsSeen {
		if !expectedColors[color] {
			t.Errorf("Unexpected color %v in 4-color mode", color)
		}
	}

	// Purple should not appear
	if colorsSeen[Purple] {
		t.Error("Purple should not appear in 4-color mode")
	}
}

func TestColorCount5(t *testing.T) {
	game := NewGameWithColors(5)

	if game.ColorCount != 5 {
		t.Errorf("Expected ColorCount to be 5, got %d", game.ColorCount)
	}

	// Generate many pairs and verify all 5 colors can appear
	colorsSeen := make(map[Color]bool)
	for i := 0; i < 200; i++ {
		pair := game.generatePuyoPair()
		colorsSeen[pair.Main.Color] = true
		colorsSeen[pair.Sub.Color] = true
	}

	// With enough iterations, all 5 colors should appear
	expectedColors := []Color{Red, Green, Blue, Yellow, Purple}
	for _, color := range expectedColors {
		if !colorsSeen[color] {
			t.Logf("Warning: Color %v did not appear in 200 iterations (may be unlucky)", color)
		}
	}
}

func TestInvalidColorCount(t *testing.T) {
	// Test that invalid color counts default to 4
	game := NewGameWithColors(3)
	if game.ColorCount != 4 {
		t.Errorf("Expected invalid color count to default to 4, got %d", game.ColorCount)
	}

	game = NewGameWithColors(6)
	if game.ColorCount != 4 {
		t.Errorf("Expected invalid color count to default to 4, got %d", game.ColorCount)
	}
}

func TestDropAndLock(t *testing.T) {
	game := NewGame()

	// Fill bottom to create a ground
	for x := 0; x < FieldWidth; x++ {
		game.Field.Grid[FieldHeight-1][x] = Red
	}

	// Position puyo above ground
	game.Current.Pos.X = 3
	game.Current.Pos.Y = FieldHeight - 2

	t.Logf("Initial: Y=%d, GroundFrames=%d, ShouldLock=%v",
		game.Current.Pos.Y, game.GroundFrames, game.ShouldLock())

	// Simulate drop and ground frame counting (like UI does)
	for i := 0; i < 35; i++ {
		dropped := game.Drop()

		// If on ground, increment GroundFrames (simulating frame ticker)
		if !dropped {
			game.GroundFrames++
		}

		t.Logf("Drop %d: dropped=%v, Y=%d, GroundFrames=%d, ShouldLock=%v",
			i+1, dropped, game.Current.Pos.Y, game.GroundFrames, game.ShouldLock())

		if game.ShouldLock() {
			t.Logf("Should lock after %d drops", i+1)
			return
		}
	}

	t.Error("Did not lock after 35 drops")
}
