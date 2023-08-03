package agentimpl

import (
	"fmt"
	"github.com/ERupshis/metrics/internal/agent/metricsgetter"
	"github.com/ERupshis/metrics/internal/agent/options"
	"github.com/go-resty/resty/v2"
	"math/rand"
	"runtime"
)

type Agent struct {
	stats  runtime.MemStats
	client *resty.Client

	opts      options.Options
	pollCount int64
}

func Create(opts options.Options) *Agent {
	return &Agent{client: resty.New(), opts: opts}
}

func CreateDefault() *Agent {
	opts := options.Options{
		Host:           "http://localhost:8080",
		ReportInterval: 10,
		PollInterval:   2,
	}
	return &Agent{client: resty.New(), opts: opts}
}

func (a *Agent) GetPollInterval() int64 {
	return a.opts.PollInterval
}

func (a *Agent) GetReportInterval() int64 {
	return a.opts.ReportInterval
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
	return a.opts.Host + "/update/gauge/" + name + "/" + fmt.Sprintf("%f", value)
}

func (a *Agent) createCounterURL(name string, value int64) string {
	return a.opts.Host + "/update/counter/" + name + "/" + fmt.Sprintf("%d", value)
}
