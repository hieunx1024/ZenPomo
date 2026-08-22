package daemon

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"
	"zenpomo/internal/core"
)

// Client sends IPC commands to the daemon.
type Client struct{}

// NewClient returns a new IPC client.
func NewClient() *Client {
	return &Client{}
}

// SendCommand executes a single command against the running daemon.
func (c *Client) SendCommand(cmd string, payload ...string) (Response, error) {
	conn, err := dialIPC()
	if err != nil {
		return Response{}, fmt.Errorf("daemon not reachable: %w", err)
	}
	defer conn.Close()

	var p string
	if len(payload) > 0 {
		p = payload[0]
	}

	req := Request{Command: cmd, Payload: p}
	data, err := json.Marshal(req)
	if err != nil {
		return Response{}, err
	}

	if _, err := conn.Write(append(data, '\n')); err != nil {
		return Response{}, err
	}

	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		var resp Response
		if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
			return Response{}, fmt.Errorf("invalid daemon response: %w", err)
		}
		return resp, nil
	}

	return Response{}, fmt.Errorf("empty response from daemon")
}

// SendTUICommand executes a command identifying the sender as the TUI interface.
func (c *Client) SendTUICommand(cmd string) (Response, error) {
	conn, err := dialIPC()
	if err != nil {
		return Response{}, fmt.Errorf("daemon not reachable: %w", err)
	}
	defer conn.Close()

	req := Request{Command: cmd, Sender: "tui"}
	data, err := json.Marshal(req)
	if err != nil {
		return Response{}, err
	}

	if _, err := conn.Write(append(data, '\n')); err != nil {
		return Response{}, err
	}

	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		var resp Response
		if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
			return Response{}, fmt.Errorf("invalid daemon response: %w", err)
		}
		return resp, nil
	}

	return Response{}, fmt.Errorf("empty response from daemon")
}

// SendConfigCommand sends a command with a Config payload.
func (c *Client) SendConfigCommand(cmd string, cfg core.Config) (Response, error) {
	conn, err := dialIPC()
	if err != nil {
		return Response{}, fmt.Errorf("daemon not reachable: %w", err)
	}
	defer conn.Close()

	req := Request{Command: cmd, Config: &cfg}
	data, err := json.Marshal(req)
	if err != nil {
		return Response{}, err
	}

	if _, err := conn.Write(append(data, '\n')); err != nil {
		return Response{}, err
	}

	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		var resp Response
		if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
			return Response{}, fmt.Errorf("invalid daemon response: %w", err)
		}
		return resp, nil
	}

	return Response{}, fmt.Errorf("empty response from daemon")
}

// IsRunning checks if the background daemon is alive.
func (c *Client) IsRunning() bool {
	resp, err := c.SendCommand(CmdPing)
	return err == nil && resp.Success
}

// EnsureDaemon ensures the background daemon process is running, starting it if necessary.
func (c *Client) EnsureDaemon() error {
	if c.IsRunning() {
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		exe = "zenpomo"
	}

	cmd := exec.Command(exe, "daemon")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	setDetachedProcess(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start background daemon: %w", err)
	}

	// Wait up to 2 seconds for daemon to initialize socket
	for i := 0; i < 20; i++ {
		time.Sleep(100 * time.Millisecond)
		if c.IsRunning() {
			return nil
		}
	}
	return fmt.Errorf("timed out waiting for daemon to start")
}

// GetSnapshot retrieves the current snapshot from the daemon.
func (c *Client) GetSnapshot() (core.SessionSnapshot, error) {
	resp, err := c.SendCommand(CmdGetStatus)
	if err != nil {
		return core.SessionSnapshot{}, err
	}
	return resp.Snapshot, nil
}

// GetConfig retrieves the current config from the daemon.
func (c *Client) GetConfig() (core.Config, error) {
	resp, err := c.SendCommand(CmdGetConfig)
	if err != nil {
		return core.DefaultConfig(), err
	}
	return resp.Config, nil
}

// UpdateConfig sends updated configuration to the daemon.
func (c *Client) UpdateConfig(cfg core.Config) error {
	_, err := c.SendConfigCommand(CmdUpdateConfig, cfg)
	return err
}
