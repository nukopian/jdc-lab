//go:build unix

package main

import (
	"os"
	"runtime"
	"syscall"
)

// peakRSS returns the most memory a finished process held at once, in
// bytes. The build constraint above keeps this file to Unix, the only
// systems with a Rusage; peak_other.go stands in elsewhere.
func peakRSS(ps *os.ProcessState) int64 {
	ru, ok := ps.SysUsage().(*syscall.Rusage)
	if !ok {
		return 0
	}
	if runtime.GOOS == "darwin" {
		return ru.Maxrss // in bytes on macOS
	}
	return ru.Maxrss << 10 // in kilobytes on Linux
}
