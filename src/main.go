// vidvault — local video gallery server
// Build:  go build -o vidvault ./src
// Run:    ./vidvault [directory] [--port 8765]
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	port := flag.Int("p", 8765, "port to listen on")
	debug := flag.Bool("d", false, "debug mode")
	disableBrowser := flag.Bool("disable-browser", false, "disable browser opening")
	flag.Parse()

	if *debug {
		fmt.Println("debug mode enabled...")
	}

	dir := "."
	if flag.NArg() > 0 {
		dir = flag.Arg(0)
	}

	root, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalf("error: cannot resolve path %q: %v", dir, err)
	}
	if fi, err := os.Stat(root); err != nil || !fi.IsDir() {
		log.Fatalf("error: %q is not a directory", root)
	}

	addr := "0.0.0.0:" + strconv.Itoa(*port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("error: cannot bind to %s: %v", addr, err)
	}

	localURL := "http://localhost:" + strconv.Itoa(*port)
	networkURL := "http://" + lanIP() + ":" + strconv.Itoa(*port)
	fmt.Printf("\n  VIDVAULT   →  %s\n", localURL)
	fmt.Printf("  network    →  %s\n", networkURL)
	fmt.Printf("  scanning   →  %s\n\n", root)
	fmt.Println("  Press Ctrl+C to stop.")
	fmt.Println()

	if *disableBrowser == false {
		go openBrowser(localURL)
	}

	srv, err := newServer(root)
	if err != nil {
		log.Fatalf("error: cannot initialize server: %v", err)
	}
	log.Fatal(http.Serve(ln, srv))
}
