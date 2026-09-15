package daemon

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

func getSocketAddress() (string, string) {
	if runtime.GOOS == "windows" {
		return "tcp", "127.0.0.1:41789"
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "unix", "/tmp/zenpomo.sock"
	}
	appDir := filepath.Join(configDir, "zenpomo")
	_ = os.MkdirAll(appDir, 0755)
	return "unix", filepath.Join(appDir, "zenpomo.sock")
}

func listenIPC() (net.Listener, error) {
	network, addr := getSocketAddress()
	if network == "unix" {
		// If a daemon is already listening, refuse to bind instead of unlinking its socket out
		// from under it: that would silently orphan the running daemon (it keeps running,
		// invisible and unreachable, while this new instance takes over) rather than fail loudly.
		if conn, err := net.DialTimeout(network, addr, 200*time.Millisecond); err == nil {
			_ = conn.Close()
			return nil, fmt.Errorf("a daemon is already listening on %s", addr)
		}
		_ = os.Remove(addr) // No live listener behind it: a stale file left by an unclean shutdown.
	}
	return net.Listen(network, addr)
}

func dialIPC() (net.Conn, error) {
	network, addr := getSocketAddress()
	return net.Dial(network, addr)
}
