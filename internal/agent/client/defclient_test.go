package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/erupshis/metrics/internal/hasher"
	"github.com/erupshis/metrics/internal/logger"
)

func TestDefaultClient_PostJSON(t *testing.T) {
	log := logger.CreateMock()

	type fields struct {
		client *http.Client
		log    logger.BaseLogger
		hash   *hasher.Hasher
	}
	type args struct {
		ctx  context.Context
		url  string
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
				client: &http.Client{},
				log:    log,
				hash:   hasher.CreateHasher("", hasher.SHA256, log),
			},
			args: args{
				ctx:  context.Background(),
				url:  "/updates/",
				body: []byte(`{"val":1}`),
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
				ctx:  context.Background(),
				url:  "/updates/",
				body: []byte(`{"val":1}`),
			},
			wantErr: false,
		},
	}
	for _, ttCommon := range tests {
		tt := ttCommon
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &DefaultClient{
				client: tt.fields.client,
				log:    tt.fields.log,
				hash:   tt.fields.hash,
			}

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			if err := c.PostJSON(tt.args.ctx, ts.URL+tt.args.url, tt.args.body); (err != nil) != tt.wantErr {
				t.Errorf("PostJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
