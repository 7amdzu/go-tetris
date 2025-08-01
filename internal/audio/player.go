package audio

import (
    "os"
    "time"

    "github.com/faiface/beep"
    "github.com/faiface/beep/mp3"
    "github.com/faiface/beep/wav"
    "github.com/faiface/beep/speaker"
)

// Play reads any .mp3 or .wav file and streams it to the default audio device.
func Play(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()

    // decode based on extension
    var streamer beep.StreamSeekCloser
    var format beep.Format
    if ext := path[len(path)-4:]; ext == ".mp3" {
        streamer, format, err = mp3.Decode(f)
    } else {
        streamer, format, err = wav.Decode(f)
    }
    if err != nil {
        return err
    }
    defer streamer.Close()

    // initialize speaker with a small buffer
    speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

    // play and block until done
    done := make(chan struct{})
    speaker.Play(beep.Seq(streamer, beep.Callback(func() {
        close(done)
    })))
    <-done

    return nil
}
