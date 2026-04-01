package main

import (
	"os/exec"
	"runtime"
)

// ── open browser ──────────────────────────────────────────────────────────────

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
	exec.Command(cmd, args...).Start()
}
