package hasher

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/erupshis/metrics/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasher_HashMsg(t *testing.T) {
	type fields struct {
		log      logger.BaseLogger
		hashType int
	}
	type args struct {
		msg []byte
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "common case",
			fields: fields{
				log:      logger.CreateMock(),
				hashType: SHA256,
			},
			args: args{
				msg: []byte("{\"some message text\"}"),
				key: "123",
			},
			want:    "b325442b7351543173366c32ad347b7f2b643e6bdabe7aa3717c819caeb9726c",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hr := &Hasher{
				log:      tt.fields.log,
				hashType: tt.fields.hashType,
				key:      tt.args.key,
			}
			got, err := hr.HashMsg(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashMsg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("HashMsg() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasher_checkRequestHash(t *testing.T) {
	type fields struct {
		log      logger.BaseLogger
		hashType int
	}
	type args struct {
		hashHeaderValue string
		hashKey         string
		body            []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "common case",
			fields: fields{
				log:      logger.CreateMock(),
				hashType: SHA256,
			},
			args: args{
				hashHeaderValue: "b325442b7351543173366c32ad347b7f2b643e6bdabe7aa3717c819caeb9726c",
				hashKey:         "123",
				body:            []byte("{\"some message text\"}"),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "empty hashKey",
			fields: fields{
				log:      logger.CreateMock(),
				hashType: SHA256,
			},
			args: args{
				hashHeaderValue: "b325442b7351543173366c32ad347b7f2b643e6bdabe7aa3717c819caeb9726c",
				hashKey:         "",
				body:            []byte("{\"some message text\"}"),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "empty hashHeaderValue",
			fields: fields{
				log:      logger.CreateMock(),
				hashType: SHA256,
			},
			args: args{
				hashHeaderValue: "",
				hashKey:         "123",
				body:            []byte("{\"some message text\"}"),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "incorrect hashHeaderValue",
			fields: fields{
				log:      logger.CreateMock(),
				hashType: SHA256,
			},
			args: args{
				hashHeaderValue: "wrong",
				hashKey:         "123",
				body:            []byte("{\"some message text\"}"),
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hr := &Hasher{
				log:      tt.fields.log,
				hashType: tt.fields.hashType,
				key:      tt.args.hashKey,
			}
			got, err := hr.checkRequestHash(tt.args.hashHeaderValue, tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkRequestHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("checkRequestHash() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasher_isRequestValid(t *testing.T) {
	type fields struct {
		log      logger.BaseLogger
		hashType int
	}
	type args struct {
		hashHeaderValue string
		hashKey         string
		buffer          []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "common case",
			fields: fields{
				hashType: SHA256,
				log:      logger.CreateMock(),
			},
			args: args{
				hashHeaderValue: "b325442b7351543173366c32ad347b7f2b643e6bdabe7aa3717c819caeb9726c",
				hashKey:         "123",
				buffer:          []byte("{\"some message text\"}"),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "empty hashHeaderValue",
			fields: fields{
				hashType: SHA256,
				log:      logger.CreateMock(),
			},
			args: args{
				hashHeaderValue: "",
				hashKey:         "123",
				buffer:          []byte("{\"some message text\"}"),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "empty hashKey",
			fields: fields{
				hashType: SHA256,
				log:      logger.CreateMock(),
			},
			args: args{
				hashHeaderValue: "b325442b7351543173366c32ad347b7f2b643e6bdabe7aa3717c819caeb9726c",
				hashKey:         "",
				buffer:          []byte("{\"some message text\"}"),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "incorrect hash",
			fields: fields{
				hashType: SHA256,
				log:      logger.CreateMock(),
			},
			args: args{
				hashHeaderValue: "wrong",
				hashKey:         "123",
				buffer:          []byte("{\"some message text\"}"),
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hr := &Hasher{
				log:      tt.fields.log,
				hashType: tt.fields.hashType,
				key:      tt.args.hashKey,
			}
			buf := bytes.Buffer{}
			buf.Write(tt.args.buffer)

			got, err := hr.isRequestValid(tt.args.hashHeaderValue, buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("isRequestValid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("isRequestValid() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasher_WriteHashHeaderInResponseIfNeed(t *testing.T) {
	type fields struct {
		log      logger.BaseLogger
		hashType int
	}
	type args struct {
		w            http.ResponseWriter
		hashKey      string
		responseBody []byte
	}
	type want struct {
		headerValue string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "common case",
			fields: fields{
				hashType: SHA256,
				log:      logger.CreateMock(),
			},
			args: args{
				w:            httptest.NewRecorder(),
				hashKey:      "123",
				responseBody: []byte("some text"),
			},
			want: want{
				headerValue: "bca3ddc838637c26b09c0804d56a81e26146931337c2b88b0db7641a9bcb3554",
			},
		},
		{
			name: "empty key",
			fields: fields{
				hashType: SHA256,
				log:      logger.CreateMock(),
			},
			args: args{
				w:            httptest.NewRecorder(),
				hashKey:      "",
				responseBody: []byte("some text"),
			},
			want: want{
				headerValue: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hr := &Hasher{
				log:      tt.fields.log,
				hashType: tt.fields.hashType,
				key:      tt.args.hashKey,
			}
			hr.WriteHashHeaderInResponseIfNeed(tt.args.w, tt.args.responseBody)
			assert.Equal(t, tt.args.w.Header().Get(hr.GetHeader()), tt.want.headerValue)
		})
	}
}

func TestHasher_Handler(t *testing.T) {
	type fields struct {
		log      logger.BaseLogger
		hashType int
	}
	type args struct {
		hashHeaderValue string
		hashKey         string
		body            []byte
	}
	type want struct {
		statusCode int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "common case",
			fields: fields{
				hashType: SHA256,
				log:      logger.CreateMock(),
			},
			args: args{
				hashHeaderValue: "b325442b7351543173366c32ad347b7f2b643e6bdabe7aa3717c819caeb9726c",
				hashKey:         "123",
				body:            []byte("{\"some message text\"}"),
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "empty key",
			fields: fields{
				hashType: SHA256,
				log:      logger.CreateMock(),
			},
			args: args{
				hashHeaderValue: "b325442b7351543173366c32ad347b7f2b643e6bdabe7aa3717c819caeb9726c",
				hashKey:         "",
				body:            []byte("{\"some message text\"}"),
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "empty hash",
			fields: fields{
				hashType: SHA256,
				log:      logger.CreateMock(),
			},
			args: args{
				hashHeaderValue: "",
				hashKey:         "123",
				body:            []byte("{\"some message text\"}"),
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "invalid hash",
			fields: fields{
				hashType: SHA256,
				log:      logger.CreateMock(),
			},
			args: args{
				hashHeaderValue: "b325442b7351543173366c32ad347b7f2b643e6bdabe7aa3717c819caeb2726c",
				hashKey:         "123",
				body:            []byte("{\"some message text\"}"),
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hr := &Hasher{
				log:      tt.fields.log,
				hashType: tt.fields.hashType,
				key:      tt.args.hashKey,
			}

			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Add(hr.GetHeader(), tt.args.hashHeaderValue)
			defer func() {
				_ = req.Body.Close()
			}()

			var buf bytes.Buffer
			buf.Write(tt.args.body)
			rc := &readCloserWrapper{
				Reader: bytes.NewReader(buf.Bytes()),
				Closer: req.Body,
			}
			req.Body = rc

			rr := httptest.NewRecorder()

			var err error
			handler := hr.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err = w.Write([]byte("correct"))
			}))

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			require.NoError(t, err)
		})
	}
}
