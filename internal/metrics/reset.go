package metrics

import (
	"time"
)

// StartDailyReset starts a goroutine that resets metrics at midnight every day
func StartDailyReset() {
	go func() {
		for {
			// Calculate time until next midnight
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			duration := next.Sub(now)

			// Wait until midnight
			time.Sleep(duration)

			// Reset metrics
			Reset()
		}
	}()
}
