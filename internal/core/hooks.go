package core

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// ExecuteHook runs a user-defined shell command asynchronously in a detached process with a timeout.
func ExecuteHook(cmdStr string, session SessionType, activeTask string) {
	cmdStr = strings.TrimSpace(cmdStr)
	if cmdStr == "" {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.CommandContext(ctx, "cmd.exe", "/c", cmdStr)
		} else {
			cmd = exec.CommandContext(ctx, "sh", "-c", cmdStr)
		}

		// Inject ZenPomo context into hook environment
		cmd.Env = append(os.Environ(),
			"ZENPOMO_SESSION="+string(session),
			"ZENPOMO_TASK="+activeTask,
		)

		_ = cmd.Run()
	}()
}
