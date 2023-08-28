package agentimpl

import (
	"math/rand"
	"runtime"

	"github.com/erupshis/metrics/internal/agent/client"
	"github.com/erupshis/metrics/internal/agent/config"
	"github.com/erupshis/metrics/internal/agent/metricsgetter"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/networkmsg"
)

type Agent struct {
	stats  runtime.MemStats
	client client.BaseClient
	logger logger.BaseLogger

	config    config.Config
	pollCount int64
}

func Create(config config.Config, logger logger.BaseLogger, client client.BaseClient) *Agent {
	return &Agent{client: client, config: config, logger: logger}
}

func CreateDefault() *Agent {
	return &Agent{client: client.CreateResty(), config: config.Default(), logger: logger.CreateLogger("Info")}
}

func (a *Agent) GetPollInterval() int64 {
	return a.config.PollInterval
}

func (a *Agent) GetReportInterval() int64 {
	return a.config.ReportInterval
}

func (a *Agent) UpdateStats() {
	a.logger.Info("agent trying to update stats.")
	runtime.ReadMemStats(&a.stats)
	a.pollCount++

	a.logger.Info("agent has completed stats posting. pollcount: %d", a.pollCount)
}

//JSON POST REQUESTS.

func (a *Agent) PostJSONStats() {
	a.logger.Info("agent trying to update stats.")
	for name, valueGetter := range metricsgetter.GaugeMetricsGetter {
		a.postJSONStat(a.createJSONGaugeMessage(name, valueGetter(&a.stats)))
	}

	a.postJSONStat(a.createJSONGaugeMessage("RandomValue", rand.Float64()))
	a.postJSONStat(a.createJSONCounterMessage("PollCount", a.pollCount))

	a.logger.Info("agent has completed stats updating.")
}

func (a *Agent) postJSONStat(body []byte) {
	a.client.PostJSON(a.config.Host+"/update/", body)
}
func (a *Agent) createJSONGaugeMessage(name string, value float64) []byte {
	return networkmsg.CreatePostUpdateMessage(networkmsg.CreateGaugeMetrics(name, value))
}

func (a *Agent) createJSONCounterMessage(name string, value int64) []byte {
	return networkmsg.CreatePostUpdateMessage(networkmsg.CreateCounterMetrics(name, value))
}
