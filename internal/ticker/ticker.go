// Package ticker provides utilities for working with time.Ticker and context in Go.
// It includes a function Run that runs a callback function at regular intervals
// specified by a time.Ticker until the provided context is canceled.
package ticker

import (
	"context"
	"time"
)

// Run executes the provided callback function at regular intervals defined by the given time.Ticker,
// until the provided context is canceled. The function continuously checks for context cancellation
// and stops the ticker to gracefully exit the loop.
//
// Parameters:
//   - ticker: The time.Ticker specifying the intervals between callback invocations.
//   - ctx: The context that, when canceled, triggers the termination of the callback execution.
//   - callback: The function to be executed at each tick of the time.Ticker.
func Run(ticker *time.Ticker, ctx context.Context, callback func()) {
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			callback()
		}
	}
}
