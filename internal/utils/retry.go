package utils

import (
	"time"
)

// Retry executes the provided function fn up to maxAttempts times with a delay between attempts.
// If fn returns nil, it returns nil immediately.
// If all attempts fail, it returns the error from the last attempt.
// This can be used to implement retry logic and acts as a basic circuit breaker mechanism.
func Retry(maxAttempts int, delay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < maxAttempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		time.Sleep(delay)
	}
	return err
}
