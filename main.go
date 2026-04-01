// vidvault — local video gallery server
// Build:  go build -o vidvault .
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
	port := flag.Int("port", 8765, "port to listen on")
	flag.IntVar(port, "p", 8765, "port to listen on (shorthand)")
	flag.Parse()

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

	addr := "127.0.0.1:" + strconv.Itoa(*port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("error: cannot bind to %s: %v", addr, err)
	}

	browserURL := "http://localhost:" + strconv.Itoa(*port)
	fmt.Printf("\n  VIDVAULT  →  %s\n", browserURL)
	fmt.Printf("  scanning   →  %s\n\n", root)
	fmt.Println("  Press Ctrl+C to stop.\n")

	go openBrowser(browserURL)

	log.Fatal(http.Serve(ln, newServer(root)))
}
