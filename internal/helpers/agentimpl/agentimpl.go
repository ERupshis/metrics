package agentimpl

import (
	"fmt"
	"github.com/ERupshis/metrics/internal/helpers/metricsgetter"
	"math/rand"
	"net/http"
	"runtime"
)

var ServerName = "http://localhost:8080"

type Agent struct {
	stats  runtime.MemStats
	client http.Client

	reportInterval int64
	pollInterval   int64
	pollCount      int64
}

func CreateAgent() *Agent {
	return &Agent{client: http.Client{}, reportInterval: 10, pollInterval: 2}
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
	req, errReq := http.NewRequest(http.MethodPost, url, nil)
	if errReq != nil {
		panic(errReq)
	}

	req.Header.Add("Content-Type", "text/plain")
	resp, errResp := a.client.Do(req)
	if errResp != nil {
		panic(errResp)
	}

	defer resp.Body.Close()
}

func createGaugeURL(name string, value float64) string {
	return ServerName + "/update/gauge/" + name + "/" + fmt.Sprintf("%f", value)
}

func createCounterURL(name string, value int64) string {
	return ServerName + "/update/counter/" + name + "/" + fmt.Sprintf("%d", value)
}
