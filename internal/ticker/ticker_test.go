package ticker

import (
	"context"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	// Set up a context with cancellation to stop the Run function
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up a ticker with a short interval for testing
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	// Set up a flag to indicate whether the callback was called
	callbackCalled := false

	// Define the callback function
	callback := func() {
		callbackCalled = true
	}

	// Run the function in a goroutine
	go Run(ticker, ctx, callback)

	// Allow some time for the Run function to run
	time.Sleep(30 * time.Millisecond)

	// Cancel the context to stop the Run function
	cancel()

	// Allow some time for the Run function to stop
	time.Sleep(10 * time.Millisecond)

	// Check if the callback was called
	if !callbackCalled {
		t.Error("Callback function was not called")
	}
}
