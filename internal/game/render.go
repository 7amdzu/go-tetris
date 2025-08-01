package game

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// PlayfieldPrimitive embeds Box for borders, sizing, focus.
type PlayfieldPrimitive struct {
	*tview.Box
	Game         *Game
	flashingRows []int // Rows that are flashing during animation
	flashOn      bool  // Whether the flash is currently on or off
}

// StatusPrimitive shows game status information (score, level, etc.)
type StatusPrimitive struct {
	*tview.Box
	Game *Game
}

// NextPiecePrimitive shows the next piece preview
type NextPiecePrimitive struct {
	*tview.Box
	Game *Game
}

// NewPlayfieldPrimitive constructs and positions the grid.
func NewPlayfieldPrimitive(g *Game, x, y, width, height int) *PlayfieldPrimitive {
	box := tview.NewBox().
		SetBorder(true).
		SetTitle(" TETRIS ")
	box.SetRect(x, y, width, height)
	return &PlayfieldPrimitive{Box: box, Game: g}
}

// NewStatusPrimitive creates a new status display box
func NewStatusPrimitive(g *Game, x, y, width, height int) *StatusPrimitive {
	box := tview.NewBox().
		SetBorder(true).
		SetTitle(" STATUS ").
		SetBorderColor(tcell.ColorBlue)
	box.SetRect(x, y, width, height)
	return &StatusPrimitive{Box: box, Game: g}
}

// NewNextPiecePrimitive creates a new next piece preview box
func NewNextPiecePrimitive(g *Game, x, y, width, height int) *NextPiecePrimitive {
	box := tview.NewBox().
		SetBorder(true).
		SetTitle(" NEXT ").
		SetBorderColor(tcell.ColorRed)
	box.SetRect(x, y, width, height)
	return &NextPiecePrimitive{Box: box, Game: g}
}

// Draw is called each frame by QueueUpdateDraw.
func (p *PlayfieldPrimitive) Draw(screen tcell.Screen) {
	// Update title with current piece info for debugging
	if p.Game.Current != nil {
		p.Box.SetTitle(" TETRIS ")
	} else {
		p.Box.SetTitle(" TETRIS ")
	}

	// Draw border & background.
	p.Box.DrawForSubclass(screen, p)
	x0, y0, width, height := p.GetInnerRect()

	// Clear the inner rectangle to prevent artifacts
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			screen.SetContent(x0+x, y0+y, ' ', nil, tcell.StyleDefault)
		}
	}

	// Draw proper content based on game state
	switch p.Game.State {
	case MainMenu:
		p.drawMainMenu(screen, x0, y0, width, height)
	case Paused:
		p.drawPausedOverlay(screen, x0, y0, width, height)
	case GameOver:
		p.drawGameOverOverlay(screen, x0, y0, width, height)
	case Animating:
		p.drawPlayfield(screen, x0, y0, width, height)
		p.drawAnimatingLines(screen, x0, y0, width, height)
	case Playing:
		p.drawPlayfield(screen, x0, y0, width, height)
	}
}

// drawMainMenu draws the welcome screen
func (p *PlayfieldPrimitive) drawMainMenu(screen tcell.Screen, x0, y0, width, height int) {
	// Clear the entire area first
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			screen.SetContent(x0+x, y0+y, ' ', nil, tcell.StyleDefault)
		}
	}

	// Draw a simple centered menu
	centerY := height / 2

	if centerY-3 >= 0 && centerY-3 < height {
		drawCenteredText(screen, x0, y0+centerY-3, width, "TETRIS", tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true))
	}

	if centerY-1 >= 0 && centerY-1 < height {
		drawCenteredText(screen, x0, y0+centerY-1, width, "Press ENTER to start", tcell.StyleDefault.Foreground(tcell.ColorYellow))
	}

	if centerY+1 >= 0 && centerY+1 < height {
		drawCenteredText(screen, x0, y0+centerY+1, width, "Arrow Keys: Move/Rotate", tcell.StyleDefault)
	}

	if centerY+2 >= 0 && centerY+2 < height {
		drawCenteredText(screen, x0, y0+centerY+2, width, "Space: Drop • ESC: Pause", tcell.StyleDefault)
	}

	if centerY+3 >= 0 && centerY+3 < height {
		drawCenteredText(screen, x0, y0+centerY+3, width, "Q: Quit", tcell.StyleDefault)
	}
}

// drawPausedOverlay draws the pause screen
func (p *PlayfieldPrimitive) drawPausedOverlay(screen tcell.Screen, x0, y0, width, height int) {
	// Draw the playfield in the background
	p.drawPlayfield(screen, x0, y0, width, height)

	// Draw semi-transparent overlay
	style := tcell.StyleDefault.Background(tcell.ColorBlack.TrueColor() & 0x80FFFFFF)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if x > width/4 && x < width*3/4 && y > height/3 && y < height*2/3 {
				// Skip the center (for text)
				continue
			}
			screen.SetContent(x0+x, y0+y, ' ', nil, style)
		}
	}

	// Draw pause message
	drawCenteredText(screen, x0, y0+height/2-1, width, "PAUSED", tcell.StyleDefault.Foreground(tcell.ColorYellow))
	drawCenteredText(screen, x0, y0+height/2+1, width, "Press P or ENTER to resume", tcell.StyleDefault)
}

// drawGameOverOverlay draws the game over screen
func (p *PlayfieldPrimitive) drawGameOverOverlay(screen tcell.Screen, x0, y0, width, height int) {
	// Draw the final playfield state in the background
	p.drawPlayfield(screen, x0, y0, width, height)

	// Draw semi-transparent overlay
	style := tcell.StyleDefault.Background(tcell.ColorBlack.TrueColor() & 0x80FFFFFF)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if x > width/4 && x < width*3/4 && y > height/3 && y < height*2/3 {
				// Skip the center (for text)
				continue
			}
			screen.SetContent(x0+x, y0+y, ' ', nil, style)
		}
	}

	// Draw game over message
	drawCenteredText(screen, x0, y0+height/2-2, width, "GAME OVER", tcell.StyleDefault.Foreground(tcell.ColorRed))
	gameOverScore := fmt.Sprintf("Score: %d", p.Game.Score)
	drawCenteredText(screen, x0, y0+height/2, width, gameOverScore, tcell.StyleDefault)
	drawCenteredText(screen, x0, y0+height/2+2, width, "Press ENTER to restart", tcell.StyleDefault)
}

// drawPlayfield draws the main game grid and active piece
func (p *PlayfieldPrimitive) drawPlayfield(screen tcell.Screen, x0, y0, width, height int) {
	// Calculate available space for the playfield
	playfieldWidth := PlayWidth * 2 // Double-width blocks
	playfieldHeight := VisibleHeight

	// Center the playfield within the available space
	startX := x0 + (width-playfieldWidth)/2
	startY := y0 + (height-playfieldHeight)/2

	// Ensure we don't go outside bounds
	if startX < x0 {
		startX = x0
	}
	if startY < y0 {
		startY = y0
	}

	// Draw the game grid
	for screenRow := 0; screenRow < playfieldHeight; screenRow++ {
		// Convert screen row to playfield row (flip Y coordinate)
		playfieldRow := playfieldHeight - 1 - screenRow

		for col := 0; col < PlayWidth; col++ {
			// Calculate screen position
			screenX := startX + col*2
			screenY := startY + screenRow

			// Skip if outside available area
			if screenX >= x0+width-1 || screenY >= y0+height {
				continue
			}

			// Get the cell value from playfield
			cellVal := p.Game.Playfield[col][playfieldRow]

			// Choose character and style
			ch, style := ' ', tcell.StyleDefault.Background(tcell.ColorBlack)
			if cellVal != 0 {
				ch = '█'
				style = tcell.StyleDefault.Foreground(ColorFor(cellVal)).Background(tcell.ColorBlack)
			}

			// Draw the block (2 characters wide)
			screen.SetContent(screenX, screenY, ch, nil, style)
			if screenX+1 < x0+width {
				screen.SetContent(screenX+1, screenY, ch, nil, style)
			}
		}
	}

	// Draw the current falling piece if present
	if p.Game.Current != nil {
		for _, b := range p.Game.Current.Blocks {
			// Calculate absolute position in playfield
			col := p.Game.Current.Position.X + b.X
			playfieldRow := p.Game.Current.Position.Y + b.Y

			// Only draw blocks that are within the visible area
			if col >= 0 && col < PlayWidth && playfieldRow >= 0 && playfieldRow < VisibleHeight {
				// Convert to screen coordinates
				screenRow := playfieldHeight - 1 - playfieldRow
				screenX := startX + col*2
				screenY := startY + screenRow

				// Skip if outside available area
				if screenX >= x0+width-1 || screenY >= y0+height {
					continue
				}

				// Draw the falling piece block
				style := tcell.StyleDefault.Foreground(p.Game.Current.Color)
				screen.SetContent(screenX, screenY, '█', nil, style)
				if screenX+1 < x0+width {
					screen.SetContent(screenX+1, screenY, '█', nil, style)
				}
			}
		}
	}
}

// drawAnimatingLines draws flashing animation for rows being cleared
func (p *PlayfieldPrimitive) drawAnimatingLines(screen tcell.Screen, x0, y0, width, height int) {
	// Toggle flash state for animation
	p.flashOn = !p.flashOn

	// Get current flashing rows from the game state
	// This should be set when entering Animating state
	if len(p.flashingRows) == 0 {
		// Use a default if not set properly
		return
	}

	// Draw flashing rows
	flashColor := tcell.ColorWhite
	if !p.flashOn {
		flashColor = tcell.ColorRed
	}

	// Draw each flashing row
	for _, playFieldRow := range p.flashingRows {
		// Skip if row is outside of the visible area
		if playFieldRow < HiddenBuffer {
			continue
		}

		// Convert playfield row to visible row index
		visRow := playFieldRow - HiddenBuffer

		// Convert to screen coordinates (inverted so 0 is at bottom)
		screenY := y0 + (VisibleHeight - 1 - visRow)

		// Draw the flashing row
		if screenY >= y0 && screenY < y0+height {
			style := tcell.StyleDefault.Foreground(flashColor)
			for col := 0; col < PlayWidth && col*2 < width; col++ {
				screen.SetContent(x0+col*2, screenY, '█', nil, style)
				if x0+col*2+1 < x0+width {
					screen.SetContent(x0+col*2+1, screenY, '█', nil, style)
				}
			}
		}
	}
}

// drawCenteredText draws text centered horizontally at the given y position
func drawCenteredText(screen tcell.Screen, x, y, width int, text string, style tcell.Style) {
	textWidth := len(text)
	startX := x + (width-textWidth)/2
	for i, r := range text {
		screen.SetContent(startX+i, y, r, nil, style)
	}
}

// Draw method for StatusPrimitive
func (s *StatusPrimitive) Draw(screen tcell.Screen) {
	// Draw border & background
	s.Box.DrawForSubclass(screen, s)
	x0, y0, width, height := s.GetInnerRect()

	// Clear the inner area
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			screen.SetContent(x0+x, y0+y, ' ', nil, tcell.StyleDefault)
		}
	}

	// Draw game status information
	currentLine := 0

	// Game State
	var stateText string
	switch s.Game.State {
	case Playing:
		stateText = "PLAYING"
	case Paused:
		stateText = "PAUSED"
	case GameOver:
		stateText = "GAME OVER"
	case MainMenu:
		stateText = "MAIN MENU"
	case Animating:
		stateText = "CLEARING"
	default:
		stateText = "UNKNOWN"
	}

	if currentLine < height {
		drawLeftAlignedText(screen, x0, y0+currentLine, width, fmt.Sprintf("State: %s", stateText), tcell.StyleDefault.Foreground(tcell.ColorYellow))
		currentLine += 2
	}

	// Score
	if currentLine < height {
		drawLeftAlignedText(screen, x0, y0+currentLine, width, fmt.Sprintf("Score: %d", s.Game.Score), tcell.StyleDefault.Foreground(tcell.ColorGreen))
		currentLine += 1
	}

	// Level
	if currentLine < height {
		drawLeftAlignedText(screen, x0, y0+currentLine, width, fmt.Sprintf("Level: %d", s.Game.Level), tcell.StyleDefault.Foreground(tcell.ColorBlue))
		currentLine += 1
	}

	// Lines Cleared
	if currentLine < height {
		drawLeftAlignedText(screen, x0, y0+currentLine, width, fmt.Sprintf("Lines: %d", s.Game.LinesCleared), tcell.StyleDefault.Foreground(tcell.ColorPurple))
		currentLine += 2
	}

	// Add key shortcuts help
	if currentLine < height {
		drawLeftAlignedText(screen, x0, y0+currentLine, width, "CONTROLS:", tcell.StyleDefault.Foreground(tcell.ColorWhite))
		currentLine += 1
	}

	controls := []string{
		"← → Move",
		"↑ Rotate",
		"↓ Soft Drop",
		"Space Drop",
		"ESC Pause",
		"Q Quit",
	}

	for _, control := range controls {
		if currentLine < height {
			drawLeftAlignedText(screen, x0, y0+currentLine, width, control, tcell.StyleDefault.Foreground(tcell.ColorGray))
			currentLine += 1
		}
	}
}

// drawLeftAlignedText draws text left-aligned at the given position
func drawLeftAlignedText(screen tcell.Screen, x, y, width int, text string, style tcell.Style) {
	for i, r := range text {
		if i >= width {
			break
		}
		screen.SetContent(x+i, y, r, nil, style)
	}
}

// Draw method for NextPiecePrimitive
func (n *NextPiecePrimitive) Draw(screen tcell.Screen) {
	// Draw border & background
	n.Box.DrawForSubclass(screen, n)
	x0, y0, width, height := n.GetInnerRect()

	// Clear the inner area
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			screen.SetContent(x0+x, y0+y, ' ', nil, tcell.StyleDefault.Background(tcell.ColorBlack))
		}
	}

	// Draw the next piece if available
	// NextQueue[0] is always the next piece that will spawn when current piece locks
	if len(n.Game.NextQueue) > 0 {
		nextPieceID := n.Game.NextQueue[0]

		// Validate piece ID
		if nextPieceID < I || nextPieceID > Z {
			return // Invalid piece ID, don't draw anything
		}

		// Update the NEXT box title
		n.Box.SetTitle(" NEXT ")
		shape := PieceShapes[nextPieceID][0]
		color := PieceColors[nextPieceID]

		// Calculate center position for the piece
		centerX := x0 + width/2
		centerY := y0 + height/2

		// Draw the piece blocks
		for y := 0; y < 4; y++ {
			for x := 0; x < 4; x++ {
				if shape[y][x] {
					// Calculate screen position (centered)
					screenX := centerX + (x-2)*2 // *2 for double-width, -2 to center 4x4 grid
					screenY := centerY + (y - 2) // -2 to center 4x4 grid

					// Only draw if within bounds
					if screenX >= x0 && screenX < x0+width-1 && screenY >= y0 && screenY < y0+height {
						style := tcell.StyleDefault.Foreground(color).Background(tcell.ColorBlack)
						screen.SetContent(screenX, screenY, '█', nil, style)
						if screenX+1 < x0+width {
							screen.SetContent(screenX+1, screenY, '█', nil, style)
						}
					}
				}
			}
		}
	}
}
