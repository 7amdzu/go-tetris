package game

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"gotetris/internal/audio"
)

const TickRate = time.Second / 60 // 60 Hz

// NewGame now accepts the TUI app
func NewGame(app *tview.Application, audioManager *audio.AudioManager) *Game {
	g := &Game{
		Playfield:    [PlayWidth][TotalHeight]int{},
		NextQueue:    make([]PieceID, 0, 7),
		State:        MainMenu,
		Level:        1,
		app:          app,
		audioManager: audioManager,
		quit:         make(chan struct{}),
		input:        make(chan *tcell.EventKey, 16),
	}
	g.gravityTicker = time.NewTicker(gravityForLevel(g.Level))
	return g
}

// Run starts the concurrent loop and blocks until exit.
func (g *Game) Run() error {
	// Set up input forwarding
	g.app.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		select {
		case g.input <- ev:
		default: // drop if buffer full
		}
		return ev
	})

	// Initialize screen, layouts, etc.
	if err := g.initScreen(); err != nil {
		return err
	}

	// Start the main loop in this goroutine
	go g.loop()

	// Run the tview application (blocks)
	return g.app.Run()
}

// loop is the select‑driven heartbeat
func (g *Game) loop() {
	ticker := time.NewTicker(TickRate)
	defer ticker.Stop()

	// Ensure the gravity ticker is initialized
	if g.gravityTicker == nil {
		g.gravityTicker = time.NewTicker(gravityForLevel(g.Level))
	}
	defer g.gravityTicker.Stop()

	needsRedraw := true // Initial redraw needed

	for {
		select {
		case <-ticker.C:
			// Only redraw if we're in Animating state (for line clear animations)
			if g.State == Animating {
				needsRedraw = true
			}

		case <-g.gravityTicker.C:
			// Strict conditions for applying gravity
			if g.State == Playing && g.Current != nil {
				// Additional validation before applying gravity
				if g.Current.Position.Y >= 0 && g.Current.Position.Y < TotalHeight {
					g.ApplyGravity()
					needsRedraw = true
				}
			}

		case ev := <-g.input:
			switch ev.Key() {
			case tcell.KeyCtrlC:
				close(g.quit)
				return
			default:
				oldState := g.State
				g.HandleInput(ev)
				// Only redraw if state actually changed or we're in a playable state
				if g.State != oldState || g.State == Playing || g.State == MainMenu || g.State == Paused {
					needsRedraw = true
				}
			}

		case <-g.quit:
			g.app.Stop()
			return
		}

		// Only queue a redraw when needed
		if needsRedraw {
			needsRedraw = false
			g.app.QueueUpdateDraw(func() {
				g.render()
			})
		}
	}
}
