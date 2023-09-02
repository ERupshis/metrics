package ticker

import (
	"context"
	"time"
)

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
