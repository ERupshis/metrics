package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/erupshis/metrics/internal/hasher"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/go-resty/resty/v2"
)

func TestRestyClient_PostJSON(t *testing.T) {
	log := logger.CreateMock()

	type fields struct {
		client *resty.Client
		log    logger.BaseLogger
		hash   *hasher.Hasher
	}
	type args struct {
		context context.Context
		url     string
		body    []byte
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
				client: resty.New(),
				log:    log,
				hash:   hasher.CreateHasher("", hasher.SHA256, log),
			},
			args: args{
				context: context.Background(),
				url:     "/updates/",
				body:    []byte(`{"val":1}`),
			},
			wantErr: false,
		},
		{
			name: "valid with hash key",
			fields: fields{
				client: resty.New(),
				log:    log,
				hash:   hasher.CreateHasher("1234", hasher.SHA256, log),
			},
			args: args{
				context: context.Background(),
				url:     "/updates/",
				body:    []byte(`{"val":1}`),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RestyClient{
				client: tt.fields.client,
				log:    tt.fields.log,
				hash:   tt.fields.hash,
			}

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			if err := c.PostJSON(tt.args.context, ts.URL+tt.args.url, tt.args.body); (err != nil) != tt.wantErr {
				t.Errorf("PostJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
