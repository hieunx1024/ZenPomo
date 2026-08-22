//go:build darwin

package audio

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func playOSWav(wavData []byte, name string) {
	tmpFile := filepath.Join(os.TempDir(), "zenpomo_"+name)
	if err := os.WriteFile(tmpFile, wavData, 0644); err == nil {
		_ = exec.Command("afplay", tmpFile).Run()
	} else {
		fmt.Print("\a")
	}
}
