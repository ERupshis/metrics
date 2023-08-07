package ticker

import "time"

func CreateWithSecondsInterval(seconds int64) *time.Ticker {
	return createInterval(time.Duration(seconds) * time.Second)
}

func createInterval(interval time.Duration) *time.Ticker {
	return time.NewTicker(interval)
}

func Run(ticker *time.Ticker, callback func()) {
	for range ticker.C {
		callback()
	}
}
