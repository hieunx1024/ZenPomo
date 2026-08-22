package daemon

import (
	"testing"
	"time"
)

func TestDaemon_IPCCommunication(t *testing.T) {
	srv, err := NewServer()
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	go func() {
		_ = srv.Start()
	}()
	defer srv.Stop()

	// Wait briefly for socket binding
	time.Sleep(100 * time.Millisecond)

	client := NewClient()
	if !client.IsRunning() {
		t.Fatalf("expected daemon to be running")
	}

	snap, err := client.GetSnapshot()
	if err != nil {
		t.Fatalf("failed to get snapshot: %v", err)
	}
	if snap.Session != "Work" {
		t.Fatalf("expected Work session, got %v", snap.Session)
	}

	// Test Toggle command
	resp, err := client.SendCommand(CmdToggle)
	if err != nil || !resp.Success {
		t.Fatalf("toggle failed: %v", err)
	}
	if resp.Snapshot.State != "Running" {
		t.Fatalf("expected State Running, got %v", resp.Snapshot.State)
	}

	// Test SetTask command
	resp, err = client.SendCommand(CmdSetTask, "Unit Test Task")
	if err != nil || !resp.Success {
		t.Fatalf("set task failed: %v", err)
	}
	if resp.Snapshot.ActiveTaskTitle != "Unit Test Task" {
		t.Fatalf("expected active task 'Unit Test Task', got %v", resp.Snapshot.ActiveTaskTitle)
	}
}
