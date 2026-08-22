package audio

import (
	"sync"
	"zenpomo/assets"
)

// Player handles cross-platform audio playback.
type Player struct {
	mu      sync.Mutex
	enabled bool
}

var (
	globalPlayer *Player
	once         sync.Once
)

// GetPlayer returns the singleton audio player.
func GetPlayer() *Player {
	once.Do(func() {
		globalPlayer = &Player{enabled: true}
	})
	return globalPlayer
}

// SetEnabled toggles sound playback.
func (p *Player) SetEnabled(enabled bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = enabled
}

// PlayWorkEnd plays the calming bell when focus time ends.
func (p *Player) PlayWorkEnd() {
	p.playWav(assets.SoundWorkEnd, "work_end.wav")
}

// PlayBreakEnd plays the crisp chime when break time ends.
func (p *Player) PlayBreakEnd() {
	p.playWav(assets.SoundBreakEnd, "break_end.wav")
}

func (p *Player) playWav(wavData []byte, name string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.enabled {
		return
	}

	go playOSWav(wavData, name)
}
