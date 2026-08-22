package daemon

import (
	"net"
	"os"
	"path/filepath"
	"runtime"
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
		_ = os.Remove(addr) // Clean up stale socket file
	}
	return net.Listen(network, addr)
}

func dialIPC() (net.Conn, error) {
	network, addr := getSocketAddress()
	return net.Dial(network, addr)
}
