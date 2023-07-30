package handlers

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func Test_splitCounterAndGaugeRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     string
		want    []string
		wantErr bool
	}{
		{
			name:    "correct counter request",
			req:     "/update/counter/someMetric/527",
			want:    []string{"", "update", "counter", "someMetric", "527"},
			wantErr: false,
		},
		{
			name:    "incorrect counter req with addition path",
			req:     "/update/counter/someMetric/527/ff",
			want:    []string{},
			wantErr: true,
		},
		{
			name:    "incorrect counter req without value",
			req:     "/update/counter/someMetric/",
			want:    []string{},
			wantErr: true,
		},
		{
			name:    "incorrect counter req without value and slash",
			req:     "/update/counter/someMetric",
			want:    []string{},
			wantErr: true,
		},
		{
			name:    "incorrect counter req with addition path",
			req:     "/update/counter",
			want:    []string{},
			wantErr: true,
		},
		{
			name:    "incorrect counter request with float value",
			req:     "/update/counter/someMetric/527.12",
			want:    []string{"", "update", "counter", "someMetric", "527.12"},
			wantErr: false,
		},
		{
			name:    "correct gauge request",
			req:     "/update/gauge/someMetric/527.12",
			want:    []string{"", "update", "gauge", "someMetric", "527.12"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := splitCounterAndGaugeRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("splitCounterAndGaugeRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitCounterAndGaugeRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInvalid(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name   string
		reqUrl string
		want   want
	}{
		{"test common invalid case", "/status", want{http.StatusBadRequest, "", ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.reqUrl, nil)
			// создаём новый Recorder
			w := httptest.NewRecorder()
			Invalid(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, res.StatusCode, tt.want.code)
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			if len(resBody) != 0 {
				assert.JSONEq(t, string(resBody), tt.want.response)
			}
			assert.Equal(t, res.Header.Get("Content-Type"), tt.want.contentType)
		})
	}
}

func TestCounter(t *testing.T) {
	type req struct {
		method string
		url    string
	}
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name string
		req  req
		want want
	}{
		{
			"common valid case",
			req{http.MethodPost, "/update/counter/someMetrics/345"},
			want{http.StatusOK, "", "text/plain; charset=utf-8"},
		},
		{
			"invalid case, wrong method type",
			req{http.MethodGet, "/update/counter/someMetrics/345"},
			want{http.StatusMethodNotAllowed, "", ""},
		},
		{
			"invalid case, wrong req structure",
			req{http.MethodPost, "/update/counter/someMetrics/345/"},
			want{http.StatusNotFound, "", ""},
		},
		{
			"invalid case, wrong req structure",
			req{http.MethodPost, "/update/counter/someMetrics/"},
			want{http.StatusNotFound, "", ""},
		},
		{
			"invalid case, wrong req structure",
			req{http.MethodPost, "/update/counter/someMetrics"},
			want{http.StatusNotFound, "", ""},
		},
		{
			"invalid case, wrong req structure",
			req{http.MethodPost, "/update/counter/"},
			want{http.StatusNotFound, "", ""},
		},
		{
			"invalid case, wrong req structure",
			req{http.MethodPost, "/update/counter//"},
			want{http.StatusNotFound, "", ""},
		},
		{
			"invalid case, wrong value type",
			req{http.MethodPost, "/update/counter/someMetrics/345.3"},
			want{http.StatusBadRequest, "", ""},
		},
		{
			"invalid case, wrong value type",
			req{http.MethodPost, "/update/counter/someMetrics/asd"},
			want{http.StatusBadRequest, "", ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.req.method, tt.req.url, nil)
			// создаём новый Recorder
			w := httptest.NewRecorder()
			Counter(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, res.StatusCode, tt.want.code)
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			if len(resBody) != 0 {
				assert.JSONEq(t, string(resBody), tt.want.response)
			}
			assert.Equal(t, res.Header.Get("Content-Type"), tt.want.contentType)
		})
	}
}

func TestGauge(t *testing.T) {
	type req struct {
		method string
		url    string
	}
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name string
		req  req
		want want
	}{
		{
			"common valid case",
			req{http.MethodPost, "/update/gauge/someMetrics/345.43"},
			want{http.StatusOK, "", "text/plain; charset=utf-8"},
		},
		{
			"invalid case, wrong method type",
			req{http.MethodGet, "/update/gauge/someMetrics/345.33"},
			want{http.StatusMethodNotAllowed, "", ""},
		},
		{
			"invalid case, wrong req structure",
			req{http.MethodPost, "/update/gauge/someMetrics/345.33/"},
			want{http.StatusNotFound, "", ""},
		},
		{
			"invalid case, wrong req structure",
			req{http.MethodPost, "/update/gauge/someMetrics/"},
			want{http.StatusNotFound, "", ""},
		},
		{
			"invalid case, wrong req structure",
			req{http.MethodPost, "/update/gauge/someMetrics"},
			want{http.StatusNotFound, "", ""},
		},
		{
			"invalid case, wrong req structure",
			req{http.MethodPost, "/update/gauge/"},
			want{http.StatusNotFound, "", ""},
		},
		{
			"invalid case, wrong req structure",
			req{http.MethodPost, "/update/gauge//"},
			want{http.StatusNotFound, "", ""},
		},
		{
			"invalid case, wrong value type",
			req{http.MethodPost, "/update/gauge/someMetrics/asd"},
			want{http.StatusBadRequest, "", ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.req.method, tt.req.url, nil)
			// создаём новый Recorder
			w := httptest.NewRecorder()
			Gauge(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, res.StatusCode, tt.want.code)
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			if len(resBody) != 0 {
				assert.JSONEq(t, string(resBody), tt.want.response)
			}
			assert.Equal(t, res.Header.Get("Content-Type"), tt.want.contentType)
		})
	}
}
