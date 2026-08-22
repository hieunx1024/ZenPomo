//go:build !windows && !darwin

package audio

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func playOSWav(wavData []byte, name string) {
	players := []string{"paplay", "pw-play", "aplay"}

	tmpFile := filepath.Join(os.TempDir(), "zenpomo_"+name)
	if err := os.WriteFile(tmpFile, wavData, 0644); err != nil {
		fmt.Print("\a")
		return
	}

	for _, player := range players {
		if path, err := exec.LookPath(player); err == nil {
			cmd := exec.Command(path, tmpFile)
			if err := cmd.Run(); err == nil {
				return
			}
		}
	}

	// Fallback to ASCII terminal bell
	fmt.Print("\a")
}
