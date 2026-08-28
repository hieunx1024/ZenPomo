package audio

import (
	"sync"
)

// AmbientSound represents the selected ambient noise background.
type AmbientSound string

const (
	AmbientNone       AmbientSound = "none"
	AmbientRain       AmbientSound = "rain"
	AmbientWhiteNoise AmbientSound = "whitenoise"
	AmbientWaves      AmbientSound = "waves"
	AmbientCoffee     AmbientSound = "coffee"
)

// AmbientPlayer manages persistent background ambiance.
type AmbientPlayer struct {
	mu      sync.Mutex
	current AmbientSound
	active  bool
}

var (
	globalAmbient *AmbientPlayer
	ambientOnce   sync.Once
)

// GetAmbientPlayer returns the singleton ambient audio controller.
func GetAmbientPlayer() *AmbientPlayer {
	ambientOnce.Do(func() {
		globalAmbient = &AmbientPlayer{
			current: AmbientNone,
			active:  false,
		}
	})
	return globalAmbient
}

// SetSound sets the ambient sound track.
func (ap *AmbientPlayer) SetSound(sound AmbientSound) {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	ap.current = sound
	if sound == AmbientNone {
		ap.active = false
	}
}

// GetSound returns the current ambient sound selection.
func (ap *AmbientPlayer) GetSound() AmbientSound {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	return ap.current
}

// Play begins the ambient loop if sound is not none.
func (ap *AmbientPlayer) Play() {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	if ap.current != AmbientNone {
		ap.active = true
	}
}

// Pause pauses the ambient loop.
func (ap *AmbientPlayer) Pause() {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	ap.active = false
}

// IsActive returns whether ambient audio is actively playing.
func (ap *AmbientPlayer) IsActive() bool {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	return ap.active
}
