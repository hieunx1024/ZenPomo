package audio

import (
	"testing"
)

func TestAmbientPlayer(t *testing.T) {
	ap := GetAmbientPlayer()
	ap.SetSound(AmbientRain)

	if ap.GetSound() != AmbientRain {
		t.Errorf("GetSound() = %v; want %v", ap.GetSound(), AmbientRain)
	}

	ap.Play()
	if !ap.IsActive() {
		t.Errorf("IsActive() = false after Play; want true")
	}

	ap.Pause()
	if ap.IsActive() {
		t.Errorf("IsActive() = true after Pause; want false")
	}

	ap.SetSound(AmbientNone)
	if ap.IsActive() {
		t.Errorf("IsActive() should be false when set to AmbientNone")
	}
}
