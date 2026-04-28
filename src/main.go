// vidvault — local video gallery server
// Build:  go build -o vidvault ./src
// Run:    ./vidvault [directory] [--port 8765]
package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

var mainLog = fileLogger("main")

func main() {
	port := flag.Int("p", 8765, "port to listen on")
	debug := flag.Bool("d", false, "debug mode")
	disableBrowser := flag.Bool("disable-browser", false, "disable browser opening")
	flag.Parse()

	configureLogging(*debug)
	logDebug(mainLog, "debug mode enabled")

	dir := "."
	if flag.NArg() > 0 {
		dir = flag.Arg(0)
	}
	logDebug(mainLog, "resolved input directory flag", "dir", dir)

	root, err := filepath.Abs(dir)
	if err != nil {
		mainLog.Error("cannot resolve path", "dir", dir, "error", err)
		os.Exit(1)
	}
	logDebug(mainLog, "resolved absolute root directory", "root", root)
	if fi, err := os.Stat(root); err != nil || !fi.IsDir() {
		mainLog.Error("path is not a directory", "root", root, "error", err)
		os.Exit(1)
	}

	addr := "0.0.0.0:" + strconv.Itoa(*port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		mainLog.Error("cannot bind to address", "address", addr, "error", err)
		os.Exit(1)
	}
	logDebug(mainLog, "tcp listener created", "address", addr)

	localURL := "http://localhost:" + strconv.Itoa(*port)
	networkURL := "http://" + lanIP() + ":" + strconv.Itoa(*port)
	fmt.Printf("\n  VIDVAULT   →  %s\n", localURL)
	fmt.Printf("  network    →  %s\n", networkURL)
	fmt.Printf("  scanning   →  %s\n\n", root)
	fmt.Println("  Press Ctrl+C to stop.")
	fmt.Println()

	if *disableBrowser == false {
		logDebug(mainLog, "launching browser opener goroutine", "url", localURL)
		go openBrowser(localURL)
	} else {
		logDebug(mainLog, "browser auto-open disabled")
	}

	srv, err := newServer(root)
	if err != nil {
		mainLog.Error("cannot initialize server", "error", err)
		os.Exit(1)
	}
	mainLog.Info("server started", "local_url", localURL, "network_url", networkURL, "root", root, "debug", *debug)
	if err := http.Serve(ln, srv); err != nil {
		mainLog.Error("http server stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}
