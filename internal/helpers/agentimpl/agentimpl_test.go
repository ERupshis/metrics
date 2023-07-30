package agentimpl

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateAgent(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"successful agent generation"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, CreateAgent())
		})
	}
}

func TestAgent_GetPollInterval(t *testing.T) {
	tests := []struct {
		name string
		want int64
	}{
		{"default pollInterval value", 2},
	}

	for _, tt := range tests {
		agent := CreateAgent()
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, agent.GetPollInterval(), tt.want)
		})
	}
}

func TestAgent_GetReportInterval(t *testing.T) {
	tests := []struct {
		name string
		want int64
	}{
		{"default reportInterval value", 10},
	}

	for _, tt := range tests {
		agent := CreateAgent()
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, agent.GetReportInterval(), tt.want)
		})
	}
}

func TestAgent_UpdateStats(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"runtime stats updates"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := CreateAgent()
			agent.UpdateStats()
			var pollCountOld int64 = agent.pollCount
			agent.UpdateStats()
			assert.NotEqual(t, pollCountOld, agent.pollCount)
		})
	}
}

func Test_createGaugeUrl(t *testing.T) {
	type args struct {
		name  string
		value float64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"generation gauge post request URL", args{"testGauge", 123.456}, ServerName + "/update/gauge/testGauge/123.456000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, createGaugeUrl(tt.args.name, tt.args.value), "createGaugeUrl(%v, %v)", tt.args.name, tt.args.value)
		})
	}
}

func Test_createCounterUrl(t *testing.T) {
	type args struct {
		name  string
		value int64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"generation counter post request URL", args{"testCounter", 123}, ServerName + "/update/counter/testCounter/123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, createCounterUrl(tt.args.name, tt.args.value), "createCounterUrl(%v, %v)", tt.args.name, tt.args.value)
		})
	}
}
