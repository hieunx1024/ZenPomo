//go:build !windows

package daemon

import (
	"os/exec"
	"syscall"
)

func setDetachedProcess(cmd *exec.Cmd) {
	SetDetachedProcess(cmd)
}

// SetDetachedProcess sets platform-specific detached process attributes.
func SetDetachedProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
}
