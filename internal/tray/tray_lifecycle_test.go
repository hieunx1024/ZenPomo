package tray

import (
	"testing"
	"time"
	"zenpomo/internal/daemon"
)

func TestCompleteTrayTUILifecycle(t *testing.T) {
	// Mock launch function so test does not spawn GUI windows
	oldFn := launchTUIFn
	var launchedArgs []string
	launchTUIFn = func(appArgs ...string) error {
		launchedArgs = appArgs
		return nil
	}
	defer func() { launchTUIFn = oldFn }()

	client := daemon.NewClient()

	// STEP 1: Ensure Daemon is running
	if err := client.EnsureDaemon(); err != nil {
		t.Fatalf("Failed to ensure daemon: %v", err)
	}

	// STEP 2: IS_TUI_ACTIVE when no TUI is sending heartbeats
	time.Sleep(2100 * time.Millisecond) // Ensure heartbeat timeout
	resp, err := client.SendCommand(daemon.CmdIsTUIActive)
	if err != nil {
		t.Fatalf("IS_TUI_ACTIVE command failed: %v", err)
	}
	if resp.Success {
		t.Errorf("Expected IS_TUI_ACTIVE to be false when no TUI is running, got true")
	}

	// STEP 3: Tray click 'Open TUI' when no TUI is open -> triggers launch
	launchedArgs = nil
	err = FocusOrLaunchTUI()
	if err != nil {
		t.Fatalf("FocusOrLaunchTUI error: %v", err)
	}
	if len(launchedArgs) == 0 || launchedArgs[0] != "tui" {
		t.Errorf("Expected launchNewTUI('tui'), got args=%v", launchedArgs)
	}

	// STEP 4: Simulate TUI running and sending Heartbeat
	statusResp, err := client.SendTUICommand(daemon.CmdGetStatus)
	if err != nil {
		t.Fatalf("SendTUICommand failed: %v", err)
	}
	t.Logf("TUI Heartbeat active. State=%s, Session=%s", statusResp.Snapshot.State, statusResp.Snapshot.Session)

	// Check IS_TUI_ACTIVE -> must be true
	activeResp, err := client.SendCommand(daemon.CmdIsTUIActive)
	if err != nil || !activeResp.Success {
		t.Errorf("Expected IS_TUI_ACTIVE to be true right after heartbeat, got success=%v, err=%v", activeResp.Success, err)
	}

	// STEP 5: Tray click 'Settings' while TUI is open -> switches tab to 'config' without spawning new window
	launchedArgs = nil
	err = FocusOrLaunchConfig()
	if err != nil {
		t.Fatalf("FocusOrLaunchConfig error: %v", err)
	}
	// Verify no new window was spawned
	if len(launchedArgs) > 0 {
		t.Errorf("Expected no new window spawned when TUI is active, got args=%v", launchedArgs)
	}

	// Verify TUI receives requested_mode = 'config'
	checkResp, err := client.SendTUICommand(daemon.CmdGetStatus)
	if err != nil || checkResp.RequestedMode != "config" {
		t.Errorf("Expected TUI to receive RequestedMode 'config', got '%s'", checkResp.RequestedMode)
	}

	// STEP 6: Simulate closing terminal window (Heartbeat stops)
	time.Sleep(2100 * time.Millisecond)
	timeoutResp, _ := client.SendCommand(daemon.CmdIsTUIActive)
	if timeoutResp.Success {
		t.Errorf("Expected IS_TUI_ACTIVE to be false after 2.1s timeout, got true")
	}

	// STEP 7: Tray click 'Open TUI' after window closed -> must launch fresh instance
	launchedArgs = nil
	err = FocusOrLaunchTUI()
	if err != nil {
		t.Fatalf("FocusOrLaunchTUI post-close error: %v", err)
	}
	if len(launchedArgs) == 0 || launchedArgs[0] != "tui" {
		t.Errorf("Expected launchNewTUI('tui') after close, got args=%v", launchedArgs)
	}
}
