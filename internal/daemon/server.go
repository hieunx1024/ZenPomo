package daemon

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"zenpomo/internal/audio"
	"zenpomo/internal/core"
	"zenpomo/internal/notify"
	"zenpomo/internal/storage"
)

// Server represents the background Daemon process.
type Server struct {
	mu          sync.Mutex
	timer       *core.Timer
	store       *storage.Store
	audio       *audio.Player
	listener    net.Listener
	stopChan    chan struct{}
	pendingMode string
}

// NewServer creates a new daemon instance.
func NewServer() (*Server, error) {
	store, err := storage.NewStore()
	if err != nil {
		return nil, fmt.Errorf("failed to init storage: %w", err)
	}

	cfg := store.GetConfig()
	timer := core.NewTimer(cfg)
	player := audio.GetPlayer()
	player.SetEnabled(cfg.SoundEnabled)

	srv := &Server{
		timer:    timer,
		store:    store,
		audio:    player,
		stopChan: make(chan struct{}),
	}

	timer.SetOnComplete(func(prev core.SessionType, next core.SessionType) {
		snap := timer.Snapshot()
		currentCfg := store.GetConfig()
		if prev == core.SessionWork {
			srv.audio.PlayWorkEnd()
			if currentCfg.NotificationEnable {
				notify.NotifySessionEnd(string(prev), string(next))
			}
			srv.store.IncrementPomo(snap.ActiveTaskTitle, int(currentCfg.WorkDuration.Minutes()))
		} else {
			srv.audio.PlayBreakEnd()
			if currentCfg.NotificationEnable {
				notify.NotifySessionEnd(string(prev), string(next))
			}
		}
	})

	return srv, nil
}

// Start runs the IPC server and internal 1-second ticker.
func (s *Server) Start() error {
	l, err := listenIPC()
	if err != nil {
		return fmt.Errorf("daemon is already running or socket error: %w", err)
	}
	s.listener = l

	// 1-second ticker loop
	go s.tickerLoop()

	// Handle graceful shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		s.Stop()
	}()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.stopChan:
				return nil
			default:
				continue
			}
		}
		go s.handleConnection(conn)
	}
}

// Stop cleanly terminates the daemon.
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	select {
	case <-s.stopChan:
		return
	default:
		close(s.stopChan)
		if s.listener != nil {
			_ = s.listener.Close()
		}
	}
}

func (s *Server) tickerLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.timer.Tick()
		}
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		line := scanner.Bytes()
		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			continue
		}

		resp := s.executeCommand(req)
		data, _ := json.Marshal(resp)
		_, _ = conn.Write(append(data, '\n'))
	}
}

func (s *Server) executeCommand(req Request) Response {
	s.mu.Lock()
	defer s.mu.Unlock()

	makeResp := func(success bool, msg string, mode string) Response {
		return Response{
			Success:       success,
			Message:       msg,
			Snapshot:      s.timer.Snapshot(),
			Config:        s.timer.GetConfig(),
			Stats:         s.store.GetTodayStats(),
			Tasks:         s.store.GetTasks(),
			RequestedMode: mode,
		}
	}

	switch req.Command {
	case CmdRequestConfig:
		s.pendingMode = "config"
		return makeResp(true, "", "")

	case CmdGetStatus, CmdPing:
		var mode string
		if req.Sender == "tui" {
			mode = s.pendingMode
			s.pendingMode = ""
		}
		return makeResp(true, "", mode)

	case CmdGetConfig:
		return makeResp(true, "", "")

	case CmdUpdateConfig:
		if req.Config != nil {
			s.timer.UpdateConfig(*req.Config)
			s.store.UpdateConfig(*req.Config)
			s.audio.SetEnabled(req.Config.SoundEnabled)
		}
		return makeResp(true, "", "")

	case CmdToggle:
		s.timer.Toggle()
		return makeResp(true, "", "")

	case CmdStart:
		s.timer.Start()
		return makeResp(true, "", "")

	case CmdPause:
		s.timer.Pause()
		return makeResp(true, "", "")

	case CmdReset:
		s.timer.Reset()
		return makeResp(true, "", "")

	case CmdSkip:
		s.timer.Skip()
		return makeResp(true, "", "")

	case CmdSetTask:
		s.timer.SetTask(req.Payload)
		return makeResp(true, "", "")

	case CmdToggleSound:
		enabled := s.timer.ToggleSound()
		s.audio.SetEnabled(enabled)
		cfg := s.timer.GetConfig()
		s.store.UpdateConfig(cfg)
		return makeResp(true, "", "")

	case CmdSwitchSession:
		var target core.SessionType
		switch req.Payload {
		case "short_break", "Short Break":
			target = core.SessionShortBreak
		case "long_break", "Long Break":
			target = core.SessionLongBreak
		default:
			target = core.SessionWork
		}
		s.timer.SwitchSession(target)
		s.timer.Start()
		return makeResp(true, "", "")

	case CmdStop:
		go func() {
			time.Sleep(100 * time.Millisecond)
			s.Stop()
			os.Exit(0)
		}()
		return Response{Success: true, Message: "Daemon stopping"}

	default:
		return makeResp(false, "Unknown command", "")
	}
}
