package retryer

import (
	"context"
	"time"
)

var defIntervals = []int{1, 3, 5}

func RetryCallWithTimeout(ctx context.Context, intervals []int, callback func(context.Context) error) error {
	var err error

	if intervals == nil {
		intervals = defIntervals
	}

	for _, interval := range intervals {
		ctxWithTime, cancel := context.WithTimeout(ctx, time.Duration(interval)*time.Second)
		err = callback(ctxWithTime)
		cancel()
		if err == nil {
			return nil
		}

	}

	return err
}
