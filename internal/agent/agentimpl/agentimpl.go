package agentimpl

import (
	"math/rand"
	"runtime"

	"github.com/erupshis/metrics/internal/agent/config"
	"github.com/erupshis/metrics/internal/agent/metricsgetter"
	"github.com/erupshis/metrics/internal/compressor"
	"github.com/erupshis/metrics/internal/networkmsg"
	"github.com/go-resty/resty/v2"
)

type Agent struct {
	stats  runtime.MemStats
	client *resty.Client

	config    config.Config
	pollCount int64
}

func Create(config config.Config) *Agent {
	return &Agent{client: resty.New(), config: config}
}

func CreateDefault() *Agent {
	return &Agent{client: resty.New(), config: config.Default()}
}

func (a *Agent) GetPollInterval() int64 {
	return a.config.PollInterval
}

func (a *Agent) GetReportInterval() int64 {
	return a.config.ReportInterval
}

func (a *Agent) UpdateStats() {
	runtime.ReadMemStats(&a.stats)
	a.pollCount++
}

//JSON POST REQUESTS.

func (a *Agent) PostJSONStats() {
	for name, valueGetter := range metricsgetter.GaugeMetricsGetter {
		a.postJSONStat(a.createJSONGaugeMessage(name, valueGetter(&a.stats)))
	}

	a.postJSONStat(a.createJSONGaugeMessage("RandomValue", rand.Float64()))
	a.postJSONStat(a.createJSONCounterMessage("PollCount", a.pollCount))
}

func (a *Agent) postJSONStat(body []byte) {
	compressedBody, _ := compressor.GzipCompress(body)
	//TODD: add logger

	_, _ = a.client.R().
		SetHeader("Content-Type", "application/json").
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
