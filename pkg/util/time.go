package util

import "time"

func MinTimeDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
