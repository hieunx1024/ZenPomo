package daemon

import "zenpomo/internal/core"

// Command constants for IPC.
const (
	CmdGetStatus     = "GET_STATUS"
	CmdGetConfig     = "GET_CONFIG"
	CmdUpdateConfig  = "UPDATE_CONFIG"
	CmdToggle        = "TOGGLE"
	CmdStart         = "START"
	CmdPause         = "PAUSE"
	CmdReset         = "RESET"
	CmdSkip          = "SKIP"
	CmdSetTask       = "SET_TASK"
	CmdToggleSound   = "TOGGLE_SOUND"
	CmdToggleTUI     = "TOGGLE_TUI"
	CmdSwitchSession = "SWITCH_SESSION"
	CmdRequestConfig = "REQUEST_CONFIG"
	CmdPing          = "PING"
	CmdStop          = "STOP"
)

// Request defines the JSON IPC request envelope.
type Request struct {
	Command string       `json:"command"`
	Payload string       `json:"payload,omitempty"`
	Config  *core.Config `json:"config,omitempty"`
	Sender  string       `json:"sender,omitempty"`
}

// Response defines the JSON IPC response envelope.
type Response struct {
	Success       bool                 `json:"success"`
	Message       string               `json:"message,omitempty"`
	Snapshot      core.SessionSnapshot `json:"snapshot"`
	Config        core.Config          `json:"config"`
	RequestedMode string               `json:"requested_mode,omitempty"`
}
