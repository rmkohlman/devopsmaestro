package buildlog

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// lockedWriter serialises writes to the shared lumberjack file. A single Write
// call writes all bytes atomically with respect to other goroutines — so
// callers that emit one Write per line (the standard fmt.Fprintf pattern)
// never see interleaved bytes within a line.
type lockedWriter struct {
	mu *sync.Mutex
	w  io.Writer
}

func (l *lockedWriter) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.Write(p)
}

// startupSweep removes log files in dir whose mtime is older than maxAgeDays.
// Lumberjack only enforces MaxAge on rotation events, so a long-quiet system
// would otherwise retain stale files forever. This sweep closes that gap.
//
// The sweep is conservative: it only considers regular files (no symlinks,
// no subdirectories) that have a ".log" extension or contain ".log." (gzipped
// rotated backups produced by lumberjack).
func startupSweep(dir string, maxAgeDays int) error {
	cutoff := time.Now().AddDate(0, 0, -maxAgeDays)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if !e.Type().IsRegular() {
			continue
		}
		name := e.Name()
		if name == "latest.log" || name == "latest.log.tmp" {
			continue
		}
		if !strings.HasSuffix(name, ".log") &&
			!strings.Contains(name, ".log.") {
			continue
		}
		full := filepath.Join(dir, name)
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(full)
		}
	}
	return nil
}
