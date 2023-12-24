package client

import (
	"github.com/erupshis/metrics/internal/networkmsg"
)

func createJSONGaugeMessage(name string, value float64) []byte {
	return networkmsg.CreatePostUpdateMessage(networkmsg.CreateGaugeMetrics(name, value))
}

func createJSONCounterMessage(name string, value int64) []byte {
	return networkmsg.CreatePostUpdateMessage(networkmsg.CreateCounterMetrics(name, value))
}
