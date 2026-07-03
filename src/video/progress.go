package video

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var scanProgressMu sync.Mutex
var scanProgressLast time.Time

// PrintScanProgress prints a throttled carriage-return progress line to stdout.
// Updates at most every 25 videos or 500ms.
func PrintScanProgress(found, hashed, cached int) {
	scanProgressMu.Lock()
	defer scanProgressMu.Unlock()
	now := time.Now()
	if found > 1 && found%25 != 0 && now.Sub(scanProgressLast) < 500*time.Millisecond {
		return
	}
	scanProgressLast = now
	fmt.Fprintf(os.Stdout, "\r  scanning   %d videos found (%d hashed, %d cached)   ", found, hashed, cached)
}

// PrintScanComplete prints a final scan summary line to stdout.
func PrintScanComplete(count int, elapsed time.Duration) {
	scanProgressMu.Lock()
	defer scanProgressMu.Unlock()
	fmt.Fprintf(os.Stdout, "\r  scan complete   %d videos in %s\n", count, elapsed.Round(time.Millisecond))
}
