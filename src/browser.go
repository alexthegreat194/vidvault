package main

import (
	"net"
	"os/exec"
	"runtime"
)

var browserLog = fileLogger("browser")

// ── open browser ──────────────────────────────────────────────────────────────

// lanIP returns the preferred outbound LAN IP address by probing a UDP
// connection (no data is sent). Falls back to "localhost" on error.
func lanIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		logDebug(browserLog, "failed to determine lan ip, falling back to localhost", "error", err)
		return "localhost"
	}
	defer conn.Close()
	ip := conn.LocalAddr().(*net.UDPAddr).IP.String()
	logDebug(browserLog, "determined lan ip", "ip", ip)
	return ip
}

// openBrowser launches url in the system default browser. The call is
// non-blocking; failures (e.g. no browser available) are silently ignored.
func openBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd, args = "open", []string{url}
	case "windows":
		cmd, args = "cmd", []string{"/c", "start", url}
	default:
		cmd, args = "xdg-open", []string{url}
	}
	logDebug(browserLog, "opening browser", "command", cmd, "url", url)
	if err := exec.Command(cmd, args...).Start(); err != nil {
		browserLog.Error("failed to open browser", "command", cmd, "url", url, "error", err)
	}
}
