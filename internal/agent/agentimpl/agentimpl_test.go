package agentimpl

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/caarlos0/env"
	"github.com/erupshis/metrics/internal/agent/client"
	"github.com/erupshis/metrics/internal/agent/config"
	"github.com/erupshis/metrics/internal/agent/metricsgetter"
	"github.com/erupshis/metrics/internal/configutils"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type envConfig struct {
	PrivateKey  string `env:"KEY_RSA"`
	Certificate string `env:"CERT_RSA"`
}

func getEnvironments() (string, string) {
	keyRSA := "../../../rsa/key.pem"
	certRSA := "../../../rsa/cert.pem"

	var envs = envConfig{}
	err := env.Parse(&envs)
	if err != nil {
		return certRSA, keyRSA
	}

	configutils.SetEnvToParamIfNeed(&keyRSA, envs.PrivateKey)
	configutils.SetEnvToParamIfNeed(&certRSA, envs.Certificate)

	return certRSA, keyRSA
}

func TestCreateAgent(t *testing.T) {
	certRSA, _ := getEnvironments()
	tests := []struct {
		name string
	}{
		{"successful agent generation"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, CreateDefault(certRSA))
		})
	}
}

func TestAgent_GetPollInterval(t *testing.T) {
	certRSA, _ := getEnvironments()

	tests := []struct {
		name string
		want time.Duration
	}{
		{"default pollInterval value", 2 * time.Second},
	}

	for _, tt := range tests {
		agent := CreateDefault(certRSA)
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, agent.GetPollInterval(), tt.want)
		})
	}
}

func TestAgent_GetReportInterval(t *testing.T) {
	certRSA, _ := getEnvironments()

	tests := []struct {
		name string
		want time.Duration
	}{
		{"default reportInterval value", 10 * time.Second},
	}

	for _, tt := range tests {
		agent := CreateDefault(certRSA)
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, agent.GetReportInterval(), tt.want)
		})
	}
}

func TestAgent_UpdateStats(t *testing.T) {
	certRSA, _ := getEnvironments()

	tests := []struct {
		name string
	}{
		{"runtime stats updates"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := CreateDefault(certRSA)
			agent.UpdateStats()
			pollCountOld := agent.pollCount.Load()
			agent.UpdateStats()
			assert.NotEqual(t, pollCountOld, agent.pollCount.Load())
		})
	}
}

func TestAgent_createJSONGaugeMessage(t *testing.T) {
	certRSA, _ := getEnvironments()

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

	a := CreateDefault(certRSA)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.want, string(a.createJSONGaugeMessage(tt.args.name, tt.args.value)))
		})
	}
}

func TestAgent_createJSONCounterMessage(t *testing.T) {
	certRSA, _ := getEnvironments()

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

	a := CreateDefault(certRSA)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.want, string(a.createJSONCounterMessage(tt.args.name, tt.args.value)))
		})
	}
}

func TestAgent_UpdateExtraStats(t *testing.T) {
	type fields struct {
		extraStats metricsgetter.ExtraStats
		logger     logger.BaseLogger
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "valid",
			fields: fields{
				extraStats: metricsgetter.ExtraStats{Data: make(map[string]float64)},
				logger:     logger.CreateMock(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent{
				extraStats:      tt.fields.extraStats,
				extraStatsMutex: sync.RWMutex{},
				logger:          tt.fields.logger,
			}
			a.UpdateExtraStats()
			assert.Equal(t, 3, len(a.extraStats.Data))
		})
	}
}

func TestAgent_postJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockBaseClient(ctrl)
	gomock.InOrder(
		mockClient.EXPECT().PostJSON(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		mockClient.EXPECT().PostJSON(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("test err")),
	)

	type fields struct {
		logger logger.BaseLogger
		config config.Config
		client client.BaseClient
	}
	type args struct {
		ctx  context.Context
		body []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid",
			fields: fields{
				logger: logger.CreateMock(),
				config: config.Config{
					Host: "/",
				},
				client: mockClient,
			},
			args: args{
				ctx:  context.Background(),
				body: []byte{},
			},
			wantErr: false,
		},
		{
			name: "client returns err",
			fields: fields{
				logger: logger.CreateMock(),
				config: config.Config{
					Host: "/",
				},
				client: mockClient,
			},
			args: args{
				ctx:  context.Background(),
				body: []byte{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent{
				client: tt.fields.client,
				logger: tt.fields.logger,
				config: tt.fields.config,
			}

			if err := a.postJSON(tt.args.ctx, tt.args.body); (err != nil) != tt.wantErr {
				t.Errorf("postJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgent_postBatchJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockBaseClient(ctrl)
	gomock.InOrder(
		mockClient.EXPECT().PostJSON(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		mockClient.EXPECT().PostJSON(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("test err")),
	)

	type fields struct {
		logger logger.BaseLogger
		config config.Config
		client client.BaseClient
	}
	type args struct {
		ctx  context.Context
		body []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid",
			fields: fields{
				logger: logger.CreateMock(),
				config: config.Config{
					Host: "/",
				},
				client: mockClient,
			},
			args: args{
				ctx:  context.Background(),
				body: []byte{},
			},
			wantErr: false,
		},
		{
			name: "client returns err",
			fields: fields{
				logger: logger.CreateMock(),
				config: config.Config{
					Host: "/",
				},
				client: mockClient,
			},
			args: args{
				ctx:  context.Background(),
				body: []byte{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent{
				client: tt.fields.client,
				logger: tt.fields.logger,
				config: tt.fields.config,
			}

			if err := a.postBatchJSON(tt.args.ctx, tt.args.body); (err != nil) != tt.wantErr {
				t.Errorf("postJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgent_PostJSONStatsBatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockBaseClient(ctrl)
	gomock.InOrder(
		mockClient.EXPECT().PostJSON(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		mockClient.EXPECT().PostJSON(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("connection err")),
	)

	type fields struct {
		client client.BaseClient
		logger logger.BaseLogger
		config config.Config
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "valid",
			fields: fields{
				logger: logger.CreateMock(),
				config: config.Config{
					Host: "/",
				},
				client: mockClient,
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: assert.NoError,
		},
		{
			name: "error from http client",
			fields: fields{
				logger: logger.CreateMock(),
				config: config.Config{
					Host: "/",
				},
				client: mockClient,
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Create(tt.fields.config, tt.fields.logger, tt.fields.client)
			a.UpdateStats()
			a.UpdateExtraStats()
			tt.wantErr(t, a.PostJSONStatsBatch(tt.args.ctx), fmt.Sprintf("PostJSONStatsBatch(%v)", tt.args.ctx))
		})
	}
}

func TestAgent_PostJSONStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockBaseClient(ctrl)
	gomock.InOrder(
		mockClient.EXPECT().PostJSON(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(64),
	)

	type fields struct {
		client client.BaseClient
		logger logger.BaseLogger
		config config.Config
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "valid",
			fields: fields{
				logger: logger.CreateMock(),
				config: config.Config{
					Host: "/",
				},
				client: mockClient,
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: assert.NoError,
		},
		{
			name: "error from http client",
			fields: fields{
				logger: logger.CreateMock(),
				config: config.Config{
					Host: "/",
				},
				client: mockClient,
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Create(tt.fields.config, tt.fields.logger, tt.fields.client)
			a.UpdateStats()
			a.UpdateExtraStats()
			a.PostJSONStats(tt.args.ctx)
		})
	}
}
