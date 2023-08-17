package controllers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/server/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type req struct {
	method string
	url    string
}
type want struct {
	code        int
	response    string
	contentType string
}
type test struct {
	name string
	req  req
	want want
}

func runTests(t *testing.T, tests *[]test, ts *httptest.Server) {
	for _, tt := range *tests {
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

func TestBaseController(t *testing.T) {
	cfg := config.Parse()

	log, err := logger.CreateRequest(cfg.LogLevel)
	if err != nil {
		panic(err)
	}
	//defer log.Sync()

	ts := httptest.NewServer(CreateBase(cfg, log).Route())
	defer ts.Close()

	badRequestTests := []test{
		//badRequestHandler
		{
			"post invalid path",
			req{http.MethodPost, "/update/count/fg/dfgdfg/dfg"},
			want{http.StatusBadRequest, "", ""},
		},
		{
			"post invalid path",
			req{http.MethodPost, "/sdf"},
			want{http.StatusBadRequest, "", ""},
		},
		{
			"get invalid path",
			req{http.MethodGet, "/update/count/sdfgdf/dfgdfg/gg"},
			want{http.StatusBadRequest, "", ""},
		},
		{
			"get invalid path",
			req{http.MethodGet, "/sdf"},
			want{http.StatusBadRequest, "", ""},
		},
		//{
		//	"post invalid path",
		//	req{http.MethodPost, "/sdf/sfdg/"},
		//	want{http.StatusBadRequest, "", ""},
		//},
		//{
		//	"get invalid path",
		//	req{http.MethodGet, "/sdf/dfsg"},
		//	want{http.StatusBadRequest, "", ""},
		//},
		//{
		//	"get invalid path",
		//	req{http.MethodGet, "/update/dfsg"},
		//	want{http.StatusBadRequest, "", ""},
		//},
	}
	runTests(t, &badRequestTests, ts)

	missingNameTests := []test{
		{
			"post update counter valid path",
			req{http.MethodPost, "/update/count/"},
			want{http.StatusNotFound, "", ""},
		},
		{
			"get update counter valid path",
			req{http.MethodGet, "/update/count/"},
			want{http.StatusNotFound, "", ""},
		},
		{
			"post value counter valid path",
			req{http.MethodPost, "/value/count/"},
			want{http.StatusNotFound, "", ""},
		},
		{
			"get value counter valid path",
			req{http.MethodGet, "/value/count/"},
			want{http.StatusNotFound, "", ""},
		},
		{
			"post update gauge valid path",
			req{http.MethodPost, "/update/gauge/"},
			want{http.StatusNotFound, "", ""},
		},
		{
			"get update gauge valid path",
			req{http.MethodGet, "/update/gauge/"},
			want{http.StatusNotFound, "", ""},
		},
		{
			"post value gauge valid path",
			req{http.MethodPost, "/value/gauge/"},
			want{http.StatusNotFound, "", ""},
		},
		{
			"get value gauge valid path",
			req{http.MethodGet, "/value/gauge/"},
			want{http.StatusNotFound, "", ""},
		},
	}
	runTests(t, &missingNameTests, ts)

	counterTests := []test{
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
			want{http.StatusMethodNotAllowed, "", ""},
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
	}
	runTests(t, &counterTests, ts)

	gaugeTests := []test{
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
			want{http.StatusMethodNotAllowed, "", ""},
		},
		{
			"gauge post invalid case(missing name)",
			req{http.MethodPost, "/update/gauge/"},
			want{http.StatusNotFound, "", ""},
		},
		{
			"gauge get valid case",
			req{http.MethodGet, "/value/gauge/someMetrics"},
			want{http.StatusOK, "345", "text/plain; charset=utf-8"},
		},
		{
			"gauge get valid case",
			req{http.MethodGet, "/value/gauge/someMetrics/"},
			want{http.StatusOK, "345", "text/plain; charset=utf-8"},
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
			want{http.StatusOK, "345", "text/plain; charset=utf-8"},
		},
		{
			"gauge post valid case(new value)",
			req{http.MethodPost, "/update/gauge/someMetricsNew/533227.036"},
			want{http.StatusOK, "", "text/plain; charset=utf-8"},
		},
		{
			"gauge get valid case(new value)",
			req{http.MethodGet, "/value/gauge/someMetricsNew"},
			want{http.StatusOK, "533227.036", "text/plain; charset=utf-8"},
		},
		{
			"gauge post valid case(new value 2)",
			req{http.MethodPost, "/update/gauge/someMetricsNew2/533227.030"},
			want{http.StatusOK, "", "text/plain; charset=utf-8"},
		},
		{
			"gauge get valid case(new value2)",
			req{http.MethodGet, "/value/gauge/someMetricsNew2"},
			want{http.StatusOK, "533227.03", "text/plain; charset=utf-8"},
		},
	}
	runTests(t, &gaugeTests, ts)
}
