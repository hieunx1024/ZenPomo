//go:build windows

package audio

import (
	"syscall"
	"unsafe"
)

func playOSWav(wavData []byte, name string) {
	winmm := syscall.NewLazyDLL("winmm.dll")
	procPlaySound := winmm.NewProc("PlaySoundW")

	const (
		SND_ASYNC     = 0x0001
		SND_NODEFAULT = 0x0002
		SND_MEMORY    = 0x0004
	)

	if len(wavData) > 0 {
		_, _, _ = procPlaySound.Call(
			uintptr(unsafe.Pointer(&wavData[0])),
			0,
			uintptr(SND_ASYNC|SND_MEMORY|SND_NODEFAULT),
		)
	}
}
