package timeutils

import "time"

// Call the f function every once every duration. Close the channel to stop the scheduler
func Every(duration time.Duration, f func()) chan int {
	quit := make(chan int)
	ticker := time.NewTicker(duration)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				f()
			case <-quit:
				return
			}
		}
	}()
	return quit
}
