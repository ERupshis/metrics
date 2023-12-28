package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/erupshis/metrics/internal/hasher"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/networkmsg"
	"github.com/erupshis/metrics/internal/rsa"
	"github.com/stretchr/testify/assert"
)

const certRSA = "../../../rsa/cert.pem"

func TestDefaultClient_PostJSON(t *testing.T) {
	log := logger.CreateMock()

	encoder, err := rsa.CreateEncoder(certRSA)
	assert.NoError(t, err, "rsa encoder create error")

	type fields struct {
		client *http.Client
		log    logger.BaseLogger
		hash   *hasher.Hasher
	}
	type args struct {
		ctx    context.Context
		url    string
		metric []networkmsg.Metric
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
				client: &http.Client{},
				log:    log,
				hash:   hasher.CreateHasher("", hasher.SHA256, log),
			},
			args: args{
				ctx:    context.Background(),
				url:    "/updates/",
				metric: []networkmsg.Metric{networkmsg.CreateCounterMetrics("val", 1)},
			},
			wantErr: false,
		},
		{
			name: "valid with hash key",
			fields: fields{
				client: &http.Client{},
				log:    log,
				hash:   hasher.CreateHasher("1234", hasher.SHA256, log),
			},
			args: args{
				ctx:    context.Background(),
				url:    "/updates/",
				metric: []networkmsg.Metric{networkmsg.CreateCounterMetrics("val", 1)},
			},
			wantErr: false,
		},
	}
	for _, ttCommon := range tests {
		tt := ttCommon
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			c := &DefaultClient{
				client:  tt.fields.client,
				log:     tt.fields.log,
				hash:    tt.fields.hash,
				encoder: encoder,
				host:    ts.URL,
			}

			defer ts.Close()

			if err := c.Post(tt.args.ctx, tt.args.metric); (err != nil) != tt.wantErr {
				t.Errorf("PostJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
