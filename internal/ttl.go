package internal

import "time"

const Infinity = time.Duration(1<<63 - 1) // 290 years

var GetDefaultTTL = func(provided time.Duration) time.Duration {
	if provided > 0 {
		return provided
	}
	return time.Hour
}
