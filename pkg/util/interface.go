package util

import "time"

// Elapsed returns stopwatch.
// you get elapsed time since you called this if you call stopwatch
func Elapsed() func() time.Duration {
	start := time.Now()
	return func() time.Duration {
		return time.Since(start)
	}
}
