package game

import (
	"math/rand/v2"
	"time"
)

// refillBag shuffles all 7 pieces.
// Updated to use math/rand/v2 because apparently rand.Seed() is so 2023
func (g *Game) refillBag() {
	bag := []PieceID{I, O, T, J, L, S, Z}
	rand.Shuffle(len(bag), func(i, j int) {
		bag[i], bag[j] = bag[j], bag[i]
	})
	g.NextQueue = append(g.NextQueue, bag...)
}

// spawnNext creates Current from NextQueue.
func (g *Game) spawnNext() {
	// Ensure we have at least 2 pieces in the queue (current + next preview)
	if len(g.NextQueue) < 2 {
		g.refillBag()
	}

	// Get the next piece ID and update the queue
	pid := g.NextQueue[0]
	g.NextQueue = g.NextQueue[1:]

	// Ensure we still have pieces for the next preview after removing current piece
	if len(g.NextQueue) < 1 {
		g.refillBag()
	}

	// Build blocks from shape[0] (initial rotation state)
	var blocks [4]Point
	idx := 0
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			if PieceShapes[pid][0][y][x] {
				blocks[idx] = Point{X: x, Y: y}
				idx++
			}
		}
	}

	// Determine spawn position
	// Center horizontally in the playfield
	spawnX := PlayWidth/2 - 2
	if spawnX < 0 {
		spawnX = 0
	}

	// Spawn at the top of the visible playfield, but not outside bounds
	// Leave some space from the very top to avoid immediate collision
	spawnY := VisibleHeight + 2 // Start in hidden buffer area (Y=22)

	// Create the new piece
	g.Current = &Piece{
		ID:            pid,
		RotationState: 0,
		Color:         PieceColors[pid],
		Position:      Point{X: spawnX, Y: spawnY},
		Blocks:        blocks,
	}

	// Check if the spawn position is valid
	if g.checkCollision() {
		// If spawn position is invalid, try moving up a bit
		for attempts := 0; attempts < 3; attempts++ {
			g.Current.Position.Y++
			if !g.checkCollision() {
				break
			}
		}

		// If still colliding after attempts, trigger game over
		if g.checkCollision() {
			g.State = GameOver
			return
		}
	}

	// Check if the newly spawned piece can move down at all
	// If it can't move down immediately (resting on locked pieces),
	// it will be locked on the next gravity tick, which is correct behavior

	// Reset rotation tracking for the new piece
	g.LastMoveWasRotation = false
}

// Call this when transitioning into Playing state.
func (g *Game) StartGame() {
	// Reset the playfield
	g.Playfield = [PlayWidth][TotalHeight]int{}

	// Reset game state
	g.Score, g.Level, g.LinesCleared = 0, 1, 0
	g.Current = nil // Clear any existing piece

	// Reset the piece queue and ensure we have enough pieces
	g.NextQueue = g.NextQueue[:0]
	g.refillBag()
	// Ensure we have enough pieces for current + next preview
	if len(g.NextQueue) < 2 {
		g.refillBag()
	}

	// Set state to Playing first
	g.State = Playing

	// Reset and restart gravity timer for the new level
	if g.gravityTicker != nil {
		g.gravityTicker.Stop()
	}
	g.gravityTicker = time.NewTicker(gravityForLevel(g.Level))

	// Now spawn the first piece
	g.spawnNext()
}
