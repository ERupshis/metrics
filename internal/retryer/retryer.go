package retryer

import (
	"context"
	"time"

	"github.com/erupshis/metrics/internal/logger"
)

var defIntervals = []int{1, 3, 5}

func RetryCallWithTimeout(ctx context.Context, log logger.BaseLogger, intervals []int, repeatableErrors []error, callback func(context.Context) error) error {
	var err error

	if intervals == nil {
		intervals = defIntervals
	}

	attempt := 0
	for _, interval := range intervals {
		ctxWithTime, cancel := context.WithTimeout(ctx, time.Duration(interval)*time.Second)
		err = callback(ctxWithTime)
		if err == nil {
			cancel()
			return nil
		}

		sleep(ctxWithTime, interval)

		attempt++
		if log != nil {
			log.Info("attempt '%d' to postJSON failed with error: %v", attempt, err)
		}

		if !canRetryCall(err, repeatableErrors) {
			cancel()
			break
		}
		cancel()
	}

	return err
}

func sleep(ctx context.Context, seconds int) {
	currentTime := time.Now()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if time.Since(currentTime) > time.Duration(seconds)*time.Second {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func canRetryCall(err error, repeatableErrors []error) bool {
	if repeatableErrors == nil {
		return true
	}

	canRetry := false
	for _, repeatableError := range repeatableErrors {
		if err.Error() == repeatableError.Error() {
			canRetry = true
		}
	}

	return canRetry
}
