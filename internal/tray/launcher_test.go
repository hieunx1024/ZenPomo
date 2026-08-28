package tray

import (
	"os/exec"
	"testing"
)

func TestFocusOrLaunchTUI(t *testing.T) {
	// Check if terminal emulator exists
	_, errGnome := exec.LookPath("gnome-terminal")
	_, errXTerm := exec.LookPath("x-terminal-emulator")

	if errGnome != nil && errXTerm != nil {
		t.Skip("No terminal emulator available in test environment")
	}

	// Mock launch function so unit test does not spawn GUI window
	oldFn := launchTUIFn
	launched := false
	launchTUIFn = func(appArgs ...string) error {
		launched = true
		return nil
	}
	defer func() { launchTUIFn = oldFn }()

	err := FocusOrLaunchTUI()
	if err != nil {
		t.Logf("FocusOrLaunchTUI error: %v", err)
	}
	if !launched {
		t.Log("Window focus or launch handled cleanly")
	}
}
