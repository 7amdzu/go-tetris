package main

import (
	"flag"
	"log"

	"gotetris/internal/audio"
	"gotetris/internal/game"

	"github.com/rivo/tview"
)

func main() {
	musicPath := flag.String("music", "assets/music.mp3", "Path to MP3/WAV soundtrack")
	loopMusic := flag.Bool("loop", true, "Loop background music")
	noMusic := flag.Bool("no-music", false, "Disable music entirely")
	flag.Parse()

	var mgr *audio.AudioManager
	if !*noMusic {
		mgr = audio.NewManager(*musicPath)
		if mgr != nil {
			mgr.SetLooping(*loopMusic)
			mgr.Play()
		}
	}

	app := tview.NewApplication()
	g := game.NewGame(app, mgr)

	if err := g.Run(); err != nil {
		log.Fatalf("Game crashed: %v", err)
	}
}
