package audio

import (
    "fmt"
    "os"
    "sync"
    "time"

    "github.com/faiface/beep"
    "github.com/faiface/beep/mp3"
    "github.com/faiface/beep/wav"
    "github.com/faiface/beep/speaker"
)

type AudioManager struct {
    mu        sync.Mutex
    streamer  beep.StreamSeekCloser
    format    beep.Format
    loop      bool
    disabled  bool
    done      chan struct{}
    filepath  string
}

func NewManager(path string) *AudioManager {
    mgr := &AudioManager{filepath: path, done: make(chan struct{})}

    f, err := os.Open(path)
    if err != nil {
        fmt.Printf("audio disabled: cannot open file: %v\n", err)
        mgr.disabled = true
        return mgr
    }
    defer f.Close()

    var s beep.StreamSeekCloser
    var format beep.Format
    if ext := path[len(path)-4:]; ext == ".mp3" {
        s, format, err = mp3.Decode(f)
    } else {
        s, format, err = wav.Decode(f)
    }
    if err != nil {
        fmt.Printf("audio disabled: decode error: %v\n", err)
        mgr.disabled = true
        return mgr
    }

    if err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)); err != nil {
        fmt.Printf("audio disabled: speaker init: %v\n", err)
        mgr.disabled = true
        return mgr
    }

    mgr.streamer = s
    mgr.format = format
    return mgr
}

// Play starts playback in a goroutine. No‑op if disabled.
func (m *AudioManager) Play() {
    if m.disabled {
        return
    }
    m.mu.Lock()
    defer m.mu.Unlock()

    // rewind
    m.streamer.Seek(0)
    var seq beep.Streamer = m.streamer
    if m.loop {
        seq = beep.Loop(-1, m.streamer)
    }
    speaker.Play(beep.Seq(seq, beep.Callback(func() {
        close(m.done)
    })))
}

// Stop stops playback immediately.
func (m *AudioManager) Stop() {
    if m.disabled {
        return
    }
    speaker.Clear()
    m.done = make(chan struct{}) // reset
}

// SetLooping toggles looping.
func (m *AudioManager) SetLooping(loop bool) {
    m.mu.Lock()
    m.loop = loop
    m.mu.Unlock()
}
