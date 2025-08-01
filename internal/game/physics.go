package game

import (
	"time"
)

// --- Lock Delay & Piece Locking ------------------------------------------------

const LockDelay = time.Millisecond * 500

// lockPiece is called after a HardDrop or when gravity collides.
// td: + or when the next row is taken(already locked)
func (g *Game) lockPiece() {
	// Safety check - should never happen but prevents crashes
	if g.Current == nil {
		return
	}

	// Paint blocks into playfield
	p := g.Current
	for _, b := range p.Blocks {
		// Calculate absolute position in playfield
		x, y := p.Position.X+b.X, p.Position.Y+b.Y

		// Only add blocks that are in the valid playfield area
		if x >= 0 && x < PlayWidth && y >= 0 && y < TotalHeight {
			g.Playfield[x][y] = int(p.ID)
		}
	}

	// Detect, clear lines & score
	cleared := g.clearLines(p)
	g.updateScore(cleared, p)

	// Level up?
	if g.LinesCleared/10+1 > g.Level {
		g.Level = g.LinesCleared/10 + 1
		g.adjustGravity()
	}

	// T-spin detection can be implemented here if needed
	// Currently just tracking for future scoring enhancements
	_ = p.ID == T && g.LastMoveWasRotation

	// Clear the current piece reference
	g.Current = nil

	// Spawn next piece
	g.spawnNext()

	// spawnNext now handles collision checking and game over internally
}

// --- Line Clearing & Animation ------------------------------------------------

// clearLines returns number of lines cleared by the last lock.
func (g *Game) clearLines(p *Piece) int {
	// Check all rows where the piece has blocks
	rows := make(map[int]struct{})
	for _, b := range p.Blocks {
		rows[p.Position.Y+b.Y] = struct{}{}
	}

	// Find rows that are completely filled
	var toClear []int
	for row := range rows {
		// Skip rows that are outside the visible playing area
		if row < 0 || row >= VisibleHeight {
			continue
		}

		full := true
		for x := 0; x < PlayWidth; x++ {
			if g.Playfield[x][row] == 0 {
				full = false
				break
			}
		}
		if full {
			toClear = append(toClear, row)
		}
	}

	if len(toClear) == 0 {
		return 0
	}

	// Animate the clearing process
	g.State = Animating
	frames := 6
	for i := 0; i < frames; i++ {
		g.app.QueueUpdateDraw(func() { g.renderClearing(toClear, i%2 == 0) })
		time.Sleep(time.Millisecond * 80)
	}

	// Actually remove the rows from bottom to top to avoid overwriting
	// Sort toClear from low to high for proper clearing
	for _, row := range toClear {
		// Shift all rows above the cleared row down within the visible area
		for y := row; y < VisibleHeight-1; y++ {
			for x := 0; x < PlayWidth; x++ {
				g.Playfield[x][y] = g.Playfield[x][y+1]
			}
		}
		// Clear the top row of visible area, or pull from hidden buffer if needed
		for x := 0; x < PlayWidth; x++ {
			if VisibleHeight < TotalHeight {
				// Pull down from hidden buffer
				g.Playfield[x][VisibleHeight-1] = g.Playfield[x][VisibleHeight]
				g.Playfield[x][VisibleHeight] = 0
			} else {
				g.Playfield[x][VisibleHeight-1] = 0
			}
		}
	}
	g.State = Playing
	g.LinesCleared += len(toClear)
	return len(toClear)
}

// renderClearing overlays flashing rows during animation.
func (g *Game) renderClearing(rows []int, flash bool) {
	// Update the PlayfieldPrimitive's flashing rows data if available
	if g.playfieldView != nil {
		g.playfieldView.flashingRows = rows
		g.playfieldView.flashOn = flash
	}

	// Render the game state
	g.render()
}

// --- Scoring & Progression -----------------------------------------------------

// scoring state
var (
	basePoints = map[int]int{1: 100, 2: 300, 3: 500, 4: 800}
	comboBonus = 50
)

// updateScore handles Guideline scoring: line clears, T‑Spins, Combo, B2B.
func (g *Game) updateScore(linesCleared int, p *Piece) {
	pts := 0

	// Check for T-spin (only when the piece is a T and the last move was a rotation)
	isTspin := false
	if p.ID == T && g.detectTSpin(p) {
		isTspin = true
	}

	// Base points calculation
	if linesCleared == 0 {
		// No lines cleared, no points
		return
	} else if isTspin {
		// T-Spin base: line‑dependent (e.g., 800 for single, 1200 for double)
		// T-Spin gives higher scores compared to regular line clears
		if linesCleared == 1 {
			pts = 800 // T-Spin Single
		} else if linesCleared == 2 {
			pts = 1200 // T-Spin Double
		} else if linesCleared == 3 {
			pts = 1600 // T-Spin Triple
		}
	} else {
		// Regular line clear points from the basePoints map
		pts = basePoints[linesCleared]
	}

	// Apply level multiplier
	pts *= g.Level

	// Back-to-Back bonus (applies to Tetrises and T-Spins)
	// Only count as B2B if the previous clear was also a Tetris or T-Spin
	if (linesCleared == 4 || isTspin) && g.B2B {
		// Apply 50% bonus for back-to-back difficult clears
		pts = int(float64(pts) * 1.5)
	}

	// Update B2B flag for next clear
	// B2B continues if this was a Tetris or T-Spin, otherwise resets
	if linesCleared == 4 || isTspin {
		g.B2B = true
	} else if linesCleared > 0 {
		// Only reset B2B if some lines were cleared but it wasn't a Tetris or T-spin
		g.B2B = false
	}
	// Note: B2B status doesn't change when no lines are cleared

	// Combo system
	if linesCleared > 0 {
		// Increment combo counter for any line clear
		g.Combo++
		// Apply combo bonus starting from the 2nd consecutive clear
		if g.Combo > 1 {
			pts += (g.Combo - 1) * comboBonus * g.Level
		}
	} else {
		// Reset combo counter when no lines are cleared
		g.Combo = 0
	}

	// Add points to score
	g.Score += pts
}

// detectTSpin checks for T-spin conditions:
// 1. The piece must be a T piece
// 2. The last move was a rotation (not a shift)
// 3. At least 3 of the 4 corners around the T's center are occupied
func (g *Game) detectTSpin(p *Piece) bool {
	// T-spin detection requires the last move to be a rotation
	if !g.LastMoveWasRotation {
		return false
	}

	// Find the center of the T piece (pivot point for rotation)
	// For a standard T piece, this is at (1,1) relative to the piece origin
	cx, cy := p.Position.X+1, p.Position.Y+1

	// Define the four corners around the center
	corners := []Point{
		{cx - 1, cy - 1}, // Top-left
		{cx + 1, cy - 1}, // Top-right
		{cx - 1, cy + 1}, // Bottom-left
		{cx + 1, cy + 1}, // Bottom-right
	}

	// Count how many corners are occupied (either by a block or by being outside the playfield)
	occupied := 0
	for _, c := range corners {
		// A corner is considered occupied if:
		// 1. It's outside the playfield boundaries
		// 2. It contains a block
		if c.X < 0 || c.X >= PlayWidth || c.Y < 0 || c.Y >= TotalHeight ||
			g.Playfield[c.X][c.Y] != 0 {
			occupied++
		}
	}

	// Standard rule: A T-spin requires at least 3 corners to be occupied
	return occupied >= 3
}

// adjustGravity resets the ticker to the new speed for the current level.
func (g *Game) adjustGravity() {
	interval := gravityForLevel(g.Level)
	g.gravityTicker.Reset(interval)
}

// gravityForLevel maps level → tick interval.
func gravityForLevel(level int) time.Duration {
	switch {
	case level < 10:
		return time.Second * time.Duration(48-level*5) / 60
	case level < 20:
		return time.Second * time.Duration(28-(level-10)*2) / 60
	case level < 30:
		return time.Second * time.Duration(8-(level-20)) / 60
	default:
		return time.Second / 60
	}
}
