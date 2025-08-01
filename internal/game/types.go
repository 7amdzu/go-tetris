package game

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"gotetris/internal/audio"
)

// --- Game Constants -----------------------------------------------------------

const (
	PlayWidth     = 10
	TotalHeight   = 40
	VisibleHeight = 20

	HiddenBuffer = TotalHeight - VisibleHeight // 20 hidden rows at top
)

// --- Game States --------------------------------------------------------------

type GameState int

const (
	MainMenu GameState = iota
	Playing
	Paused
	GameOver
	Animating
)

// --- Piece Types --------------------------------------------------------------

type PieceID int

const (
	I PieceID = iota + 1
	O
	T
	J
	L
	S
	Z
)

// --- Point represents a coordinate on the grid -------------------------------

type Point struct {
	X, Y int
}

// --- Piece describes a tetromino --------------------------------------------

type Piece struct {
	ID            PieceID
	RotationState int
	Color         tcell.Color
	Position      Point
	Blocks        [4]Point
}

// --- Game represents the complete game state -------------------------------

type Game struct {
	Playfield    [PlayWidth][TotalHeight]int
	Current      *Piece
	NextQueue    []PieceID
	Score        int
	Level        int
	LinesCleared int
	B2B          bool
	Combo        int
	State        GameState

	// Game mechanics state
	LastMoveWasRotation bool // Tracks if the last move was a rotation (for T-spin detection)

	// UI/app state
	app           *tview.Application
	gravityTicker *time.Ticker
	audioManager  *audio.AudioManager
	playfieldView *PlayfieldPrimitive // Reference to the playfield view
	statusView    *StatusPrimitive    // Reference to the status view
	nextPieceView *NextPiecePrimitive // Reference to the next piece view
	quit          chan struct{}
	input         chan *tcell.EventKey
}

// --- Helper Methods --------------------------------------------------------

// ApplyGravity moves the current piece down one row if possible.
// If collision, locks the piece.
func (g *Game) ApplyGravity() {
	// Only apply gravity in Playing state and if we have a current piece
	if g.State != Playing || g.Current == nil {
		return
	}

	// Set LastMoveWasRotation to false since this is a vertical movement
	g.LastMoveWasRotation = false

	// Store original position
	originalY := g.Current.Position.Y

	// Try to move the piece down one row (decrease Y since Y=0 is bottom)
	g.Current.Position.Y--

	// Check if this position is valid
	if g.checkCollision() {
		// If collision, restore original position
		g.Current.Position.Y = originalY

		// Lock the piece only if it's at a valid position
		if originalY >= 0 && originalY < TotalHeight {
			g.lockPiece()
		} else {
			// If piece is outside valid area, trigger game over
			g.State = GameOver
		}
	}

	// Additional safety check: if piece goes below playfield, lock it
	if g.Current != nil && g.Current.Position.Y < 0 {
		g.Current.Position.Y = 0
		g.lockPiece()
	}
}

// checkCollision checks if the current piece collides with boundaries or other blocks
func (g *Game) checkCollision() bool {
	if g.Current == nil {
		return false
	}

	for _, b := range g.Current.Blocks {
		// Calculate the absolute coordinates of this block
		x := g.Current.Position.X + b.X
		y := g.Current.Position.Y + b.Y

		// Check horizontal boundaries
		if x < 0 || x >= PlayWidth {
			return true
		}

		// Check bottom boundary (Y=0 is bottom)
		if y < 0 {
			return true
		}

		// Check top boundary (Y=TotalHeight-1 is top)
		if y >= TotalHeight {
			return true
		}

		// Check collision with existing blocks
		if g.Playfield[x][y] != 0 {
			return true
		}
	}

	return false
}

// initScreen sets up the UI layout and primitives
func (g *Game) initScreen() error {
	// Create playfield primitive
	playfield := NewPlayfieldPrimitive(g, 0, 0, 0, 0)
	g.playfieldView = playfield

	// Create status/scoring box (blue box)
	statusBox := NewStatusPrimitive(g, 0, 0, 0, 0)
	g.statusView = statusBox

	// Create next piece box (red box)
	nextPieceBox := NewNextPiecePrimitive(g, 0, 0, 0, 0)
	g.nextPieceView = nextPieceBox

	// Create main layout using a simple approach
	// Use a horizontal flex to split screen into left and right sections
	mainContainer := tview.NewFlex().SetDirection(tview.FlexColumn)

	// Left section: playfield with some padding
	leftSection := tview.NewFlex().SetDirection(tview.FlexRow)
	leftSection.AddItem(tview.NewBox(), 1, 0, false) // Top padding
	leftSection.AddItem(playfield, 0, 1, true)       // Playfield takes remaining space
	leftSection.AddItem(tview.NewBox(), 1, 0, false) // Bottom padding

	// Right section: status and next piece
	rightSection := tview.NewFlex().SetDirection(tview.FlexRow)
	rightSection.AddItem(tview.NewBox(), 1, 0, false) // Top padding
	rightSection.AddItem(statusBox, 8, 0, false)      // Status box (fixed height)
	rightSection.AddItem(tview.NewBox(), 1, 0, false) // Gap
	rightSection.AddItem(nextPieceBox, 6, 0, false)   // Next piece box (fixed height)
	rightSection.AddItem(tview.NewBox(), 0, 1, false) // Bottom flexible space

	// Add sections to main container
	mainContainer.AddItem(tview.NewBox(), 2, 0, false) // Left margin
	mainContainer.AddItem(leftSection, 26, 0, false)   // Playfield section (fixed width)
	mainContainer.AddItem(tview.NewBox(), 2, 0, false) // Gap between sections
	mainContainer.AddItem(rightSection, 32, 0, false)  // Right section (fixed width)
	mainContainer.AddItem(tview.NewBox(), 0, 1, false) // Right margin (flexible)

	// Set the main container as root
	g.app.SetRoot(mainContainer, true)
	return nil
} // HandleInput processes a single input event
func (g *Game) HandleInput(ev *tcell.EventKey) {
	switch g.State {
	case MainMenu:
		// Any key to start in main menu
		if ev.Key() == tcell.KeyEnter || (ev.Key() == tcell.KeyRune && ev.Rune() == ' ') {
			g.StartGame()
		}
	case GameOver:
		// Any key to restart after game over
		if ev.Key() == tcell.KeyEnter || (ev.Key() == tcell.KeyRune && ev.Rune() == ' ') {
			g.StartGame()
		}
	case Playing:
		// Handle pause first
		if ev.Key() == tcell.KeyEscape || (ev.Key() == tcell.KeyRune && (ev.Rune() == 'p' || ev.Rune() == 'P')) {
			g.State = Paused
			return
		}

		if g.Current != nil {
			switch ev.Key() {
			case tcell.KeyLeft:
				g.moveLeft()
			case tcell.KeyRight:
				g.moveRight()
			case tcell.KeyDown:
				g.softDrop()
			case tcell.KeyUp:
				g.rotate()
			case tcell.KeyRune:
				if ev.Rune() == ' ' {
					g.hardDrop()
				}
			}
		}
	case Paused:
		// Resume from pause
		if ev.Key() == tcell.KeyEscape ||
			ev.Key() == tcell.KeyEnter ||
			(ev.Key() == tcell.KeyRune && (ev.Rune() == ' ' || ev.Rune() == 'p' || ev.Rune() == 'P')) {
			g.State = Playing
		}
	}
}

// render updates the display
func (g *Game) render() {
	// This is handled by PlayfieldPrimitive.Draw
}

// Helper move functions
func (g *Game) moveLeft() {
	// Set LastMoveWasRotation to false since this is a horizontal movement
	g.LastMoveWasRotation = false

	g.Current.Position.X--
	if g.checkCollision() {
		g.Current.Position.X++
	}
}

func (g *Game) moveRight() {
	// Set LastMoveWasRotation to false since this is a horizontal movement
	g.LastMoveWasRotation = false

	g.Current.Position.X++
	if g.checkCollision() {
		g.Current.Position.X--
	}
}

func (g *Game) softDrop() {
	// Set LastMoveWasRotation to false since this is a vertical movement
	g.LastMoveWasRotation = false

	// Store the original position in case we need to revert
	originalY := g.Current.Position.Y

	// Move down (decrease Y coordinate for downward movement)
	g.Current.Position.Y--

	// Check if this causes a collision
	if g.checkCollision() {
		// Move back to original position
		g.Current.Position.Y = originalY

		// If we can't move down due to collision with locked pieces or bottom,
		// lock the current piece
		g.lockPiece()
	}
}

func (g *Game) hardDrop() {
	// Safety check
	if g.Current == nil {
		return
	}

	// Set LastMoveWasRotation to false since this is a vertical movement
	g.LastMoveWasRotation = false

	// Track how many cells the piece drops for scoring
	dropDistance := 0

	// Store starting position for later reference
	startY := g.Current.Position.Y

	// Find drop position - move down until collision occurs
	for {
		// Try to move down one row
		g.Current.Position.Y--

		// Check if this position is valid
		if g.checkCollision() {
			// If not valid, move back up and we've found our drop position
			g.Current.Position.Y++
			break
		}

		dropDistance++

		// Safety check to prevent infinite loops
		if dropDistance > TotalHeight {
			break
		}
	}

	// Only lock if we actually moved
	if startY != g.Current.Position.Y {
		// Lock the piece in place
		g.lockPiece()
	}
}

func (g *Game) rotate() {
	oldState := g.Current.RotationState
	g.Current.RotationState = (g.Current.RotationState + 1) % 4

	// Apply rotation by rebuilding blocks
	var newBlocks [4]Point
	idx := 0
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			if PieceShapes[g.Current.ID][g.Current.RotationState][y][x] {
				newBlocks[idx] = Point{X: x, Y: y}
				idx++
			}
		}
	}

	// Save old blocks, try new blocks
	oldBlocks := g.Current.Blocks
	g.Current.Blocks = newBlocks

	// If collision, restore old state
	if g.checkCollision() {
		g.Current.RotationState = oldState
		g.Current.Blocks = oldBlocks
		// Rotation failed, so don't set LastMoveWasRotation
		return
	}

	// Rotation succeeded, set LastMoveWasRotation to true
	g.LastMoveWasRotation = true
}
