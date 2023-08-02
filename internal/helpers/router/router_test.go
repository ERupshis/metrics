package router

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter(t *testing.T) {
	ts := httptest.NewServer(Create())
	defer ts.Close()

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
		//Invalid
		{
			"post invalid path",
			req{http.MethodPost, "/update/count"},
			want{http.StatusBadRequest, "", ""},
		},
		{
			"post invalid path",
			req{http.MethodPost, "/sdf"},
			want{http.StatusBadRequest, "", ""},
		},
		{
			"get invalid path",
			req{http.MethodGet, "/update/count"},
			want{http.StatusBadRequest, "", ""},
		},
		{
			"get invalid path",
			req{http.MethodGet, "/sdf"},
			want{http.StatusBadRequest, "", ""},
		},
		//Counter
		{
			"counter post valid case",
			req{http.MethodPost, "/update/counter/someMetrics/345"},
			want{http.StatusOK, "", "text/plain; charset=utf-8"},
		},
		{
			"counter post invalid case(additional slash)",
			req{http.MethodPost, "/update/counter/someMetrics/345/"},
			want{http.StatusBadRequest, "", ""},
		},
		{
			"counter post invalid case(missing name)",
			req{http.MethodPost, "/update/counter/"},
			want{http.StatusNotFound, "", ""},
		},
		{
			"counter post invalid case(wrong method type)",
			req{http.MethodGet, "/update/counter/someMetrics/345"},
			want{http.StatusMethodNotAllowed, "", ""},
		},
		{
			"counter post invalid case(wrong value type)",
			req{http.MethodPost, "/update/counter/someMetrics/345.1"},
			want{http.StatusBadRequest, "", ""},
		},
		{
			"counter post invalid case(missing value)",
			req{http.MethodPost, "/update/counter/someMetrics/"},
			want{http.StatusBadRequest, "", ""},
		},
		{
			"counter get valid case",
			req{http.MethodGet, "/value/counter/someMetrics"},
			want{http.StatusOK, "345", "text/plain; charset=utf-8"},
		},
		{
			"counter get valid case",
			req{http.MethodGet, "/value/counter/someMetrics/"},
			want{http.StatusOK, "345", "text/plain; charset=utf-8"},
		},
		{
			"counter get valid case(missing value)",
			req{http.MethodGet, "/value/counter/missingMetrics/"},
			want{http.StatusNotFound, "", ""},
		},
		{
			"counter post valid case",
			req{http.MethodPost, "/update/counter/someMetrics/345"},
			want{http.StatusOK, "", "text/plain; charset=utf-8"},
		},
		{
			"counter get valid case(increase value)",
			req{http.MethodGet, "/value/counter/someMetrics"},
			want{http.StatusOK, "690", "text/plain; charset=utf-8"},
		},
		//Gauge
		{
			"gauge post valid case",
			req{http.MethodPost, "/update/gauge/someMetrics/345.1"},
			want{http.StatusOK, "", "text/plain; charset=utf-8"},
		},
		{
			"gauge post valid case(wrong value type int)",
			req{http.MethodPost, "/update/gauge/someMetrics/345"},
			want{http.StatusOK, "", "text/plain; charset=utf-8"},
		},
		{
			"gauge post invalid case(additional slash)",
			req{http.MethodPost, "/update/gauge/someMetrics/345.1/"},
			want{http.StatusBadRequest, "", ""},
		},
		{
			"gauge post invalid case(wrong method type)",
			req{http.MethodGet, "/update/gauge/someMetrics/345"},
			want{http.StatusMethodNotAllowed, "", ""},
		},

		{
			"gauge post invalid case(missing value)",
			req{http.MethodPost, "/update/gauge/someMetrics/"},
			want{http.StatusBadRequest, "", ""},
		},
		{
			"gauge post invalid case(missing name)",
			req{http.MethodPost, "/update/gauge/"},
			want{http.StatusNotFound, "", ""},
		},
		{
			"gauge get valid case",
			req{http.MethodGet, "/value/gauge/someMetrics"},
			want{http.StatusOK, "345.00", "text/plain; charset=utf-8"},
		},
		{
			"gauge get valid case",
			req{http.MethodGet, "/value/gauge/someMetrics/"},
			want{http.StatusOK, "345.00", "text/plain; charset=utf-8"},
		},
		{
			"gauge get valid case(missing value)",
			req{http.MethodGet, "/value/gauge/missingMetrics/"},
			want{http.StatusNotFound, "", ""},
		},
		{
			"gauge post valid case",
			req{http.MethodPost, "/update/gauge/someMetrics/345"},
			want{http.StatusOK, "", "text/plain; charset=utf-8"},
		},
		{
			"gauge get valid case(rewrite value)",
			req{http.MethodGet, "/value/gauge/someMetrics"},
			want{http.StatusOK, "345.00", "text/plain; charset=utf-8"},
		},
		// List
		{
			"valid case html list",
			req{http.MethodGet, "/"},
			want{http.StatusOK,
				"<html><body><caption>GAUGES</caption><table border = 2><tr><td>someMetrics</td><td>345.000000</td></tr></table><caption>COUNTERS</caption><table border = 2><tr><td>someMetrics</td><td>690</td></tr></body></html>",
				"text/html; charset=utf-8"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, errReq := http.NewRequest(tt.req.method, ts.URL+tt.req.url, nil)
			require.NoError(t, errReq)

			req.Header.Add("Content-Type", "text/plain")

			resp, errResp := ts.Client().Do(req)
			assert.NoError(t, errResp)
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.want.response, string(respBody))
			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
		})
	}
}
