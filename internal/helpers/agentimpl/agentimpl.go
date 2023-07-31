package agentimpl

import (
	"fmt"
	"github.com/ERupshis/metrics/internal/helpers/metricsgetter"
	"github.com/go-resty/resty/v2"
	"math/rand"
	"runtime"
)

var ServerName = "http://localhost:8080"

type Agent struct {
	stats  runtime.MemStats
	client *resty.Client

	reportInterval int64
	pollInterval   int64
	pollCount      int64
}

func Create() *Agent {
	return &Agent{client: resty.New(), reportInterval: 10, pollInterval: 2}
}

func (a *Agent) GetPollInterval() int64 {
	return a.pollInterval
}

func (a *Agent) GetReportInterval() int64 {
	return a.reportInterval
}

func (a *Agent) UpdateStats() {
	runtime.ReadMemStats(&a.stats)
	a.pollCount++
}

func (a *Agent) PostStats() {
	for name, valueGetter := range metricsgetter.GaugeMetricsGetter {
		a.postStat(createGaugeURL(name, valueGetter(&a.stats)))
	}

	a.postStat(createGaugeURL("RandomValue", rand.Float64()))
	a.postStat(createCounterURL("PollCount", a.pollCount))
}

func (a *Agent) postStat(url string) {
	_, err := a.client.R().
		SetHeader("Content-Type", "text/plain").
		Post(url)

	if err != nil {
		panic(err)
	}
}

func createGaugeURL(name string, value float64) string {
	return ServerName + "/update/gauge/" + name + "/" + fmt.Sprintf("%f", value)
}

func createCounterURL(name string, value int64) string {
	return ServerName + "/update/counter/" + name + "/" + fmt.Sprintf("%d", value)
}
