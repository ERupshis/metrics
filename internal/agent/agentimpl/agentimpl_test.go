package agentimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAgent(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"successful agent generation"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, CreateDefault())
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
		agent := CreateDefault()
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
		agent := CreateDefault()
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
			agent := CreateDefault()
			agent.UpdateStats()
			pollCountOld := agent.pollCount
			agent.UpdateStats()
			assert.NotEqual(t, pollCountOld, agent.pollCount)
		})
	}
}

func TestAgent_createJSONGaugeMessage(t *testing.T) {
	type args struct {
		name  string
		value float64
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "valid case",
			args: args{"asd", 123},
			want: []byte("{\"id\":\"asd\",\"type\":\"gauge\",\"value\":123}"),
		},
		{
			name: "valid case without value",
			args: args{name: "asd"},
			want: []byte("{\"id\":\"asd\",\"type\":\"gauge\",\"value\":0}"),
		},
	}

	a := CreateDefault()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.want, string(a.createJSONGaugeMessage(tt.args.name, tt.args.value)))
		})
	}
}

func TestAgent_createJSONCounterMessage(t *testing.T) {
	type args struct {
		name  string
		value int64
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "valid case",
			args: args{"asd", 123},
			want: []byte("{\"id\":\"asd\",\"type\":\"counter\",\"delta\":123}"),
		},
		{
			name: "valid case without value",
			args: args{name: "asd"},
			want: []byte("{\"id\":\"asd\",\"type\":\"counter\",\"delta\":0}"),
		},
	}

	a := CreateDefault()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.want, string(a.createJSONCounterMessage(tt.args.name, tt.args.value)))
		})
	}
}
