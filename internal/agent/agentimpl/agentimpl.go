package agentimpl

import (
	"math/rand"
	"runtime"

	"github.com/erupshis/metrics/internal/agent/config"
	"github.com/erupshis/metrics/internal/agent/metricsgetter"
	"github.com/erupshis/metrics/internal/compressor"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/networkmsg"
	"github.com/go-resty/resty/v2"
)

type Agent struct {
	stats  runtime.MemStats
	client *resty.Client
	logger logger.BaseLogger

	config    config.Config
	pollCount int64
}

func Create(config config.Config, logger logger.BaseLogger) *Agent {
	return &Agent{client: resty.New(), config: config, logger: logger}
}

func CreateDefault() *Agent {
	return &Agent{client: resty.New(), config: config.Default(), logger: logger.CreateLogger("Info")}
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
	compressedBody, _ := compressor.GzipCompress(body)
	//TODD: add logger

	_, _ = a.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(compressedBody).
		Post(a.config.Host + "/update/")
}
func (a *Agent) createJSONGaugeMessage(name string, value float64) []byte {
	return networkmsg.CreatePostUpdateMessage(networkmsg.CreateGaugeMetrics(name, value))
}

func (a *Agent) createJSONCounterMessage(name string, value int64) []byte {
	return networkmsg.CreatePostUpdateMessage(networkmsg.CreateCounterMetrics(name, value))
}
