// Package agentimpl collects runtime application's metrics and sends them on server via http requests.
package agentimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/erupshis/metrics/internal/agent/client"
	"github.com/erupshis/metrics/internal/agent/config"
	"github.com/erupshis/metrics/internal/agent/metricsgetter"
	"github.com/erupshis/metrics/internal/hasher"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/networkmsg"
	"github.com/erupshis/metrics/internal/rsa"
)

type Agent struct {
	stats      runtime.MemStats
	statsMutex sync.RWMutex

	pollCount atomic.Int64

	extraStats      metricsgetter.ExtraStats
	extraStatsMutex sync.RWMutex

	client client.BaseClient
	logger logger.BaseLogger
	config config.Config
}

// Create defines agent with assigned fields from params.
func Create(config config.Config, logger logger.BaseLogger, client client.BaseClient) *Agent {
	extraStats := metricsgetter.ExtraStats{Data: make(map[string]float64)}
	return &Agent{client: client, config: config, logger: logger, extraStats: extraStats}
}

// CreateDefault agent with predefined fields. Recommended to use for debug only clauses.
func CreateDefault() *Agent {
	log := logger.CreateLogger("Info")
	hashKey := ""
	extraStats := metricsgetter.ExtraStats{Data: make(map[string]float64)}

	encoder, err := rsa.CreateEncoder("../../../rsa/cert.pem")
	if err != nil {
		log.Info("create default agent: %v", err)
		return nil
	}

	return &Agent{client: client.CreateDefault(log, hasher.CreateHasher(hashKey, hasher.SHA256, log), encoder), config: config.Default(), logger: log, extraStats: extraStats}
}

// GetPollInterval returns collecting poll interval (seconds).
func (a *Agent) GetPollInterval() int64 {
	return a.config.PollInterval
}

// GetReportInterval returns send stats to server interval.
func (a *Agent) GetReportInterval() int64 {
	return a.config.ReportInterval
}

// UpdateStats reads runtime stats and increments pollCount.
func (a *Agent) UpdateStats() {
	a.logger.Info("[Agent:UpdateStats] agent trying to update stats.")

	a.statsMutex.Lock()
	runtime.ReadMemStats(&a.stats)
	a.statsMutex.Unlock()

	a.pollCount.Add(1)

	a.logger.Info("[Agent:UpdateStats] agent has completed stats updating. pollCount: %d", a.pollCount.Load())
}

// UpdateExtraStats reads additional extra stats not included in runtime.
func (a *Agent) UpdateExtraStats() {
	a.logger.Info("[Agent:UpdateExtraStats] agent trying to update stats.")
	for key, funcVal := range metricsgetter.AdditionalGaugeMetricsGetter {
		var err error
		a.extraStatsMutex.Lock()
		a.extraStats.Data[key], err = funcVal()
		a.extraStatsMutex.Unlock()
		if err != nil {
			a.logger.Info("[Agent:UpdateExtraStats] agent failed to update extra metric '%s': %v", key, err)
		}
	}
	a.logger.Info("[Agent:UpdateExtraStats] agent has completed stats posting.")
}

// PostJSONStatsBatch sends all stats in one http post request.
func (a *Agent) PostJSONStatsBatch(ctx context.Context) error {
	a.logger.Info("[Agent:PostJSONStatsBatch] agent is trying to update stats.")
	metrics := make([]networkmsg.Metric, 0)
	for name, valueGetter := range metricsgetter.GaugeMetricsGetter {
		a.statsMutex.Lock()
		metrics = append(metrics, networkmsg.CreateGaugeMetrics(name, valueGetter(&a.stats)))
		a.statsMutex.Unlock()
	}

	for name, value := range a.extraStats.Data {
		a.extraStatsMutex.RLock()
		metrics = append(metrics, networkmsg.CreateGaugeMetrics(name, value))
		a.extraStatsMutex.RUnlock()
	}

	metrics = append(metrics, networkmsg.CreateGaugeMetrics("RandomValue", rand.Float64()))
	metrics = append(metrics, networkmsg.CreateCounterMetrics("PollCount", a.pollCount.Load()))

	body, err := json.Marshal(&metrics)
	if err != nil {
		return fmt.Errorf("[Agent:PostJSONStatsBatch] failed to create request's JSON body: %w", err)
	}

	if err = a.postBatchJSON(ctx, body); err != nil {
		return fmt.Errorf("[Agent:PostJSONStatsBatch] postBatchJSON couldn't complete sending with error: %w", err)
	}
	a.logger.Info("[Agent:PostJSONStatsBatch] stats was sent.")
	return nil
}

// PostJSONStats sends all stats in split http posts request(1 request = 1 stat).
func (a *Agent) PostJSONStats(ctx context.Context) {
	a.logger.Info("[Agent:PostJSONStats] agent is trying to update stats.")

	failedPostsCount := 0
	var err error
	for name, valueGetter := range metricsgetter.GaugeMetricsGetter {
		a.statsMutex.RLock()
		err = a.postJSON(ctx, a.createJSONGaugeMessage(name, valueGetter(&a.stats)))
		a.statsMutex.RUnlock()
		if err != nil {
			failedPostsCount++
		}
	}

	for name, value := range a.extraStats.Data {
		a.extraStatsMutex.RLock()
		err = a.postJSON(ctx, a.createJSONGaugeMessage(name, value))
		a.extraStatsMutex.RUnlock()
		if err != nil {
			failedPostsCount++
		}
	}

	err = a.postJSON(ctx, a.createJSONGaugeMessage("RandomValue", rand.Float64()))
	if err != nil {
		failedPostsCount++
	}

	err = a.postJSON(ctx, a.createJSONCounterMessage("PollCount", a.pollCount.Load()))
	if err != nil {
		failedPostsCount++
	}

	a.logger.Info("[Agent:PostJSONStats] stats was sent with failed posts: %d", failedPostsCount)
}

func (a *Agent) postBatchJSON(ctx context.Context, body []byte) error {
	err := a.client.PostJSON(ctx, a.config.Host+"/updates/", body)
	if err != nil {
		a.logger.Info("[Agent:postBatchJSON] finished with error: %v", err)
	}
	return err
}

func (a *Agent) postJSON(ctx context.Context, body []byte) error {
	err := a.client.PostJSON(ctx, a.config.Host+"/update/", body)
	if err != nil {
		a.logger.Info("[Agent:postJSON] finished with error: %v", err)
	}
	return err
}
func (a *Agent) createJSONGaugeMessage(name string, value float64) []byte {
	return networkmsg.CreatePostUpdateMessage(networkmsg.CreateGaugeMetrics(name, value))
}

func (a *Agent) createJSONCounterMessage(name string, value int64) []byte {
	return networkmsg.CreatePostUpdateMessage(networkmsg.CreateCounterMetrics(name, value))
}
