package agentimpl

import (
	"fmt"
	"math/rand"
	"runtime"

	"github.com/erupshis/metrics/internal/agent/config"
	"github.com/erupshis/metrics/internal/agent/metricsgetter"
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

func (a *Agent) PostStats() {
	for name, valueGetter := range metricsgetter.GaugeMetricsGetter {
		a.postStat(a.createGaugeURL(name, valueGetter(&a.stats)))
	}

	a.postStat(a.createGaugeURL("RandomValue", rand.Float64()))
	a.postStat(a.createCounterURL("PollCount", a.pollCount))
}

func (a *Agent) postStat(url string) {
	_, err := a.client.R().
		SetHeader("Content-Type", "text/plain").
		Post(url)

	if err != nil {
		panic(err)
	}
}

func (a *Agent) createGaugeURL(name string, value float64) string {
	return a.config.Host + "/update/gauge/" + name + "/" + fmt.Sprintf("%f", value)
}

func (a *Agent) createCounterURL(name string, value int64) string {
	return a.config.Host + "/update/counter/" + name + "/" + fmt.Sprintf("%d", value)
}