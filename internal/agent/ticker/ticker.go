package ticker

import (
	"context"
	"time"
)

func CreateWithSecondsInterval(seconds int64) *time.Ticker {
	return createInterval(time.Duration(seconds) * time.Second)
}

func createInterval(interval time.Duration) *time.Ticker {
	return time.NewTicker(interval)
}

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
