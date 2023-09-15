package agentimpl

import (
	"context"
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
	log := logger.CreateLogger("Info")
	return &Agent{client: client.CreateDefault(log), config: config.Default(), logger: log}
}

func (a *Agent) GetPollInterval() int64 {
	return a.config.PollInterval
}

func (a *Agent) GetReportInterval() int64 {
	return a.config.ReportInterval
}

func (a *Agent) UpdateStats() {
	a.logger.Info("[Agent:UpdateStats] agent trying to update stats.")
	runtime.ReadMemStats(&a.stats)
	a.pollCount++

	a.logger.Info("[Agent:UpdateStats] agent has completed stats posting. pollCount: %d", a.pollCount)
}

//JSON POST REQUESTS.

func (a *Agent) PostJSONStatsBatch(ctx context.Context) {
	a.logger.Info("[Agent:PostJSONStatsBatch] agent is trying to update stats.")
	metrics := make([]networkmsg.Metric, 0)
	for name, valueGetter := range metricsgetter.GaugeMetricsGetter {
		metrics = append(metrics, networkmsg.CreateGaugeMetrics(name, valueGetter(&a.stats)))
	}

	metrics = append(metrics, networkmsg.CreateGaugeMetrics("RandomValue", rand.Float64()))
	metrics = append(metrics, networkmsg.CreateCounterMetrics("PollCount", a.pollCount))

	body, err := json.Marshal(&metrics)
	if err != nil {
		a.logger.Info("[Agent:PostJSONStatsBatch] failed to create request's JSON body.")
		return
	}

	if err = a.postBatchJSON(ctx, body); err != nil {
		a.logger.Info("[Agent:PostJSONStatsBatch] postBatchJSON couldn't complete sending with error: %v", err)
		return
	}
	a.logger.Info("[Agent:PostJSONStatsBatch] stats was sent.")
}

func (a *Agent) PostJSONStats(ctx context.Context) {
	a.logger.Info("[Agent:PostJSONStats] agent is trying to update stats.")

	failedPostsCount := 0
	var err error
	for name, valueGetter := range metricsgetter.GaugeMetricsGetter {
		err = a.postJSON(ctx, a.createJSONGaugeMessage(name, valueGetter(&a.stats)))
		if err != nil {
			failedPostsCount++
		}
	}

	err = a.postJSON(ctx, a.createJSONGaugeMessage("RandomValue", rand.Float64()))
	if err != nil {
		failedPostsCount++
	}

	err = a.postJSON(ctx, a.createJSONCounterMessage("PollCount", a.pollCount))
	if err != nil {
		failedPostsCount++
	}

	a.logger.Info("[Agent:PostJSONStats] stats was sent with failed posts: %d", failedPostsCount)
}

func (a *Agent) postBatchJSON(ctx context.Context, body []byte) error {
	return a.client.PostJSON(ctx, a.config.Host+"/updates/", body, a.config.Key)
}

func (a *Agent) postJSON(ctx context.Context, body []byte) error {
	err := a.client.PostJSON(ctx, a.config.Host+"/update/", body, a.config.Key)
	if err != nil {
		a.logger.Info("[Agent:postBatchJSON] finished with error: %v", err)
	}
	return err
}
func (a *Agent) createJSONGaugeMessage(name string, value float64) []byte {
	return networkmsg.CreatePostUpdateMessage(networkmsg.CreateGaugeMetrics(name, value))
}

func (a *Agent) createJSONCounterMessage(name string, value int64) []byte {
	return networkmsg.CreatePostUpdateMessage(networkmsg.CreateCounterMetrics(name, value))
}
