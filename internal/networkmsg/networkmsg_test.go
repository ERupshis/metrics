package networkmsg

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePostBatchValueMessage(t *testing.T) {
	delta := int64(42)
	value := 1.42

	type args struct {
		message []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []Metric
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "valid",
			args: args{
				message: []byte(`[{"id":"counter_metric","type":"counter","delta":42}, {"id":"gauge_metric","type":"gauge","value":1.42}]`),
			},
			want: []Metric{
				CreateCounterMetrics("counter_metric", 42),
				CreateGaugeMetrics("gauge_metric", 1.42),
			},
			wantErr: assert.NoError,
		},
		{
			name: "empty message",
			args: args{
				message: []byte(``),
			},
			want:    []Metric{},
			wantErr: assert.Error,
		},
		{
			name: "incorrect message",
			args: args{
				message: []byte(`{"id":"counter_metric","]`),
			},
			want:    []Metric{},
			wantErr: assert.Error,
		},
		{
			name: "error during validation",
			args: args{
				message: []byte(`[{"id":"counter_metric","type":"counter","delta":42}, {"id":"gauge_metric","type":"gauge","delta":42,"value":1.42}]`),
			},
			want: []Metric{
				CreateCounterMetrics("counter_metric", 42),
				{
					ID:    "gauge_metric",
					MType: "gauge",
					Value: &value,
					Delta: &delta,
				},
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePostBatchValueMessage(tt.args.message)
			if !tt.wantErr(t, err, fmt.Sprintf("ParsePostBatchValueMessage(%v)", tt.args.message)) {
				return
			}
			assert.Equalf(t, tt.want, got, "ParsePostBatchValueMessage(%v)", tt.args.message)
		})
	}
}

func TestParsePostValueMessage(t *testing.T) {
	delta := int64(42)
	value := 1.42

	type args struct {
		message []byte
		delta   int64
	}
	tests := []struct {
		name    string
		args    args
		want    Metric
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "valid",
			args: args{
				message: []byte(`{"id":"counter_metric","type":"counter","delta":42}`),
			},
			want: Metric{
				ID:    "counter_metric",
				MType: "counter",
				Delta: &delta,
			},
			wantErr: assert.NoError,
		},
		{
			name: "valid",
			args: args{
				message: []byte(`{"id":"gauge_metric","type":"gauge","value":1.42}`),
			},
			want: Metric{
				ID:    "gauge_metric",
				MType: "gauge",
				Value: &value,
			},
			wantErr: assert.NoError,
		},
		{
			name: "incorrect json",
			args: args{
				message: []byte(`{"id":"counter_metr}`),
			},
			want: Metric{
				ID:    "",
				MType: "",
			},
			wantErr: assert.Error,
		},
		{
			name: "two values at hte same time",
			args: args{
				message: []byte(`{"id":"gauge_metric","type":"gauge","delta":42,"value":1.42}`),
			},
			want: Metric{
				ID:    "gauge_metric",
				MType: "gauge",
				Value: &value,
				Delta: &delta,
			},
			wantErr: assert.Error,
		},
		{
			name: "without values",
			args: args{
				message: []byte(`{"id":"gauge_metric","type":"gauge"}`),
			},
			want: Metric{
				ID:    "gauge_metric",
				MType: "gauge",
			},
			wantErr: assert.Error,
		},
		{
			name: "without type",
			args: args{
				message: []byte(`{"id":"gauge_metric"}`),
			},
			want: Metric{
				ID: "gauge_metric",
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePostValueMessage(tt.args.message)
			if !tt.wantErr(t, err, fmt.Sprintf("ParsePostValueMessage(%v)", tt.args.message)) {
				return
			}
			assert.Equalf(t, tt.want, got, "ParsePostValueMessage(%v)", tt.args.message)
		})
	}
}

func Test_isMetricValid(t *testing.T) {
	delta := int64(42)
	value := 1.42

	type args struct {
		m Metric
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "valid counter",
			args: args{
				m: CreateCounterMetrics("counter_metric", 42),
			},
			want:    true,
			wantErr: assert.NoError,
		},
		{
			name: "valid gauge",
			args: args{
				m: CreateGaugeMetrics("gauge_metric", 1.42),
			},
			want:    true,
			wantErr: assert.NoError,
		},
		{
			name: "missing name",
			args: args{
				m: Metric{
					ID:    "",
					MType: "counter",
					Delta: &delta,
				},
			},
			want:    false,
			wantErr: assert.Error,
		},
		{
			name: "missing name",
			args: args{
				m: Metric{
					ID:    "counter_metric",
					MType: "",
					Delta: &delta,
				},
			},
			want:    false,
			wantErr: assert.Error,
		},
		{
			name: "two values",
			args: args{
				m: Metric{
					ID:    "counter_metric",
					MType: "counter",
					Delta: &delta,
					Value: &value,
				},
			},
			want:    false,
			wantErr: assert.Error,
		},
		{
			name: "without values",
			args: args{
				m: Metric{
					ID:    "counter_metric",
					MType: "counter",
				},
			},
			want:    false,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isMetricValid(tt.args.m)
			if !tt.wantErr(t, err, fmt.Sprintf("isMetricValid(%v)", tt.args.m)) {
				return
			}
			assert.Equalf(t, tt.want, got, "isMetricValid(%v)", tt.args.m)
		})
	}
}

func TestCreatePostUpdateMessage(t *testing.T) {
	type args struct {
		data Metric
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "valid",
			args: args{
				data: CreateCounterMetrics("counter_metric", 42),
			},
			want: []byte(`{"id":"counter_metric","type":"counter","delta":42}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, CreatePostUpdateMessage(tt.args.data), "CreatePostUpdateMessage(%v)", tt.args.data)
		})
	}
}
