// Package retryer provides functions to recall some action with defined interval in case of error.
package retryer

import (
	"context"
	"time"

	"github.com/erupshis/metrics/internal/logger"
)

// defIntervals default intervals to retry action.
var defIntervals = []int{1, 3, 5}

// RetryCallWithTimeout retries the execution of the specified callback function
// considering time intervals until successful completion or context expiration.
// The function takes a context, a base logger, an array of waiting intervals (in seconds),
// an array of repeatable errors, and a callback function to be retried. In case of an error
// occurring during the callback execution, the function will be retried within the specified intervals.
// Parameters:
//   - ctx: Execution context providing cancellation and timeout.
//   - log: Base logger for logging information about retry attempts.
//   - intervals: Array of time intervals (in seconds) between retry attempts.
//     If intervals is nil, default values will be used.
//   - repeatableErrors: Array of errors for which retry attempts are considered valid.
//   - callback: Callback function to be retried in case of an error.
// Returns an error that occurred in the last attempt or nil in case of successful execution.

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

		<-ctxWithTime.Done()

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

// canRetryCall checks whether the given error is considered retryable
// based on the provided array of repeatable errors.
// If repeatableErrors is nil, the function considers all errors as retryable.
// Parameters:
//   - err: The error to be checked for retryability.
//   - repeatableErrors: Array of errors that are considered retryable.
//
// Returns true if the error is retryable, otherwise returns false.
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
