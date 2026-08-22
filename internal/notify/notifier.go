package notify

import (
	"os"
	"path/filepath"
	"sync"
	"zenpomo/assets"

	"github.com/gen2brain/beeep"
)

var (
	iconPathOnce sync.Once
	cachedIcon   string
)

// getTempIcon extracts the embedded tomato icon to a temp file for OS notification systems.
func getTempIcon() string {
	iconPathOnce.Do(func() {
		tmpDir := os.TempDir()
		target := filepath.Join(tmpDir, "zenpomo_icon.png")
		if err := os.WriteFile(target, assets.IconTomato, 0644); err == nil {
			cachedIcon = target
		}
	})
	return cachedIcon
}

// Send displays an OS native desktop notification.
func Send(title, message string) {
	go func() {
		icon := getTempIcon()
		_ = beeep.Notify(title, message, icon)
	}()
}

// NotifySessionEnd sends appropriate notifications based on the completed session.
func NotifySessionEnd(completedSession string, nextSession string) {
	if completedSession == "Work" {
		Send("🍅 Focus Session Complete!", "Great job! Time for a "+nextSession+". Step away and stretch!")
	} else {
		Send("⚡ Break Finished!", "Ready to dive back into deep focus? Let's start the next task!")
	}
}
