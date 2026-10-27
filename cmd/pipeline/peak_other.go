//go:build !unix

package main

import "os"

// peakRSS returns 0: only Unix reports a process's peak memory this way.
func peakRSS(*os.ProcessState) int64 { return 0 }
