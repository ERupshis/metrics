package agentimpl

import (
	"encoding/json"
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
	return &Agent{client: client.CreateDefault(), config: config.Default(), logger: logger.CreateLogger("Info")}
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

func (a *Agent) PostJSONStatsBatch() error {
	a.logger.Info("[Agent:PostJSONStatsBatch] Agent is trying to update stats.")
	metrics := make([]networkmsg.Metric, 0, 0)
	for name, valueGetter := range metricsgetter.GaugeMetricsGetter {
		metrics = append(metrics, networkmsg.CreateGaugeMetrics(name, valueGetter(&a.stats)))
	}

	metrics = append(metrics, networkmsg.CreateGaugeMetrics("RandomValue", rand.Float64()))
	metrics = append(metrics, networkmsg.CreateCounterMetrics("PollCount", a.pollCount))

	body, err := json.Marshal(&metrics)
	if err != nil {
		a.logger.Info("[Agent:PostJSONStatsBatch] Failed to create request's JSON body.")
		return err
	}

	a.postBatchJSON(body)
	a.logger.Info("[Agent:PostJSONStatsBatch] Stats was sent.")
	return nil
}

func (a *Agent) PostJSONStats() {
	a.logger.Info("[Agent:PostJSONStats] Agent is trying to update stats.")
	for name, valueGetter := range metricsgetter.GaugeMetricsGetter {
		a.postJSON(a.createJSONGaugeMessage(name, valueGetter(&a.stats)))
	}

	a.postJSON(a.createJSONGaugeMessage("RandomValue", rand.Float64()))
	a.postJSON(a.createJSONCounterMessage("PollCount", a.pollCount))

	a.logger.Info("agent has completed stats updating.")
}

func (a *Agent) postBatchJSON(body []byte) {
	a.client.PostJSON(a.config.Host+"/updates/", body)
}

func (a *Agent) postJSON(body []byte) {
	a.client.PostJSON(a.config.Host+"/update/", body)
}
func (a *Agent) createJSONGaugeMessage(name string, value float64) []byte {
	return networkmsg.CreatePostUpdateMessage(networkmsg.CreateGaugeMetrics(name, value))
}

func (a *Agent) createJSONCounterMessage(name string, value int64) []byte {
	return networkmsg.CreatePostUpdateMessage(networkmsg.CreateCounterMetrics(name, value))
}
