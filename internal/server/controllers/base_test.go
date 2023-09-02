package controllers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/erupshis/metrics/internal/compressor"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/networkmsg"
	"github.com/erupshis/metrics/internal/server/config"
	"github.com/erupshis/metrics/internal/server/memstorage"
	"github.com/erupshis/metrics/internal/server/memstorage/storagemngr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// JSON HANDLERS.
type testJSON struct {
	name string
	req  reqJSON
	want wantJSON
}

type reqJSON struct {
	method string
	url    string
	body   string
}
type wantJSON struct {
	code        int
	contentType string
	body        string
}

func TestJSONCounterBaseController(t *testing.T) {
	cfg := config.Config{
		Host:     "localhost:8080",
		LogLevel: "Info",
	}

	log, err := logger.CreateZapLogger(cfg.LogLevel)
	if err != nil {
		panic(err)
	}
	//defer log.Sync()

	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(storageManager)

	ts := httptest.NewServer(CreateBase(cfg, log, storage).Route())
	defer ts.Close()

	var val1 int64 = 123
	//var val2 int64 = 456

	counterTests := []testJSON{
		{
			"counter post without value",
			reqJSON{
				http.MethodPost,
				"/update/",
				string(networkmsg.CreatePostUpdateMessage(
					networkmsg.Metrics{
						ID:    "asd",
						MType: "counter",
					})),
			},
			wantJSON{
				http.StatusOK, "application/json",
				"{\"id\":\"asd\",\"type\":\"counter\",\"delta\":0}"},
		},
		{
			"counter post valid case",
			reqJSON{
				http.MethodPost,
				"/update/",
				string(networkmsg.CreatePostUpdateMessage(
					networkmsg.Metrics{
						ID:    "asd",
						MType: "counter",
						Delta: &val1,
					})),
			},
			wantJSON{
				http.StatusOK, "application/json",
				"{\"id\":\"asd\",\"type\":\"counter\",\"delta\":123}"},
		},
		{
			"counter get valid case",
			reqJSON{
				http.MethodPost,
				"/value/",
				string(networkmsg.CreatePostUpdateMessage(
					networkmsg.Metrics{
						ID:    "asd",
						MType: "counter",
						Delta: &val1,
					})),
			},
			wantJSON{
				http.StatusOK, "application/json",
				"{\"id\":\"asd\",\"type\":\"counter\",\"delta\":123}"},
		},
		{
			"counter post one more time case",
			reqJSON{
				http.MethodPost,
				"/update/",
				string(networkmsg.CreatePostUpdateMessage(
					networkmsg.Metrics{
						ID:    "asd",
						MType: "counter",
						Delta: &val1,
					})),
			},
			wantJSON{
				http.StatusOK, "application/json",
				"{\"id\":\"asd\",\"type\":\"counter\",\"delta\":246}"},
		},
		{
			"counter get increased valid case",
			reqJSON{
				http.MethodPost,
				"/value/",
				string(networkmsg.CreatePostUpdateMessage(
					networkmsg.Metrics{
						ID:    "asd",
						MType: "counter",
						Delta: &val1,
					})),
			},
			wantJSON{
				http.StatusOK, "application/json",
				"{\"id\":\"asd\",\"type\":\"counter\",\"delta\":246}"},
		},
		{
			"counter get missing value",
			reqJSON{
				http.MethodPost,
				"/value/",
				string(networkmsg.CreatePostUpdateMessage(
					networkmsg.Metrics{
						ID:    "asds",
						MType: "counter",
					})),
			},
			wantJSON{
				http.StatusNotFound, "text/plain; charset=utf-8",
				"invalid counter name\n"},
		},
	}
	runJSONTests(t, &counterTests, ts)

}

func TestJSONGaugeBaseController(t *testing.T) {
	cfg := config.Config{
		Host:     "localhost:8080",
		LogLevel: "Info",
	}

	log, err := logger.CreateZapLogger(cfg.LogLevel)
	if err != nil {
		panic(err)
	}
	//defer log.Sync()
	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(storageManager)

	ts := httptest.NewServer(CreateBase(cfg, log, storage).Route())
	defer ts.Close()

	var float1 float64 = 123
	var float2 = 123.23
	gaugeTests := []testJSON{
		{
			"gauge post without value",
			reqJSON{
				http.MethodPost,
				"/update/",
				string(networkmsg.CreatePostUpdateMessage(
					networkmsg.Metrics{
						ID:    "asd",
						MType: "gauge",
					})),
			},
			wantJSON{
				http.StatusOK, "application/json",
				"{\"id\":\"asd\",\"type\":\"gauge\",\"value\":0}"},
		},
		{
			"gauge post valid case",
			reqJSON{
				http.MethodPost,
				"/update/",
				string(networkmsg.CreatePostUpdateMessage(
					networkmsg.Metrics{
						ID:    "asd",
						MType: "gauge",
						Value: &float1,
					})),
			},
			wantJSON{
				http.StatusOK, "application/json",
				"{\"id\":\"asd\",\"type\":\"gauge\",\"value\":123}"},
		},
		{
			"gauge get valid case",
			reqJSON{
				http.MethodPost,
				"/value/",
				string(networkmsg.CreatePostUpdateMessage(
					networkmsg.Metrics{
						ID:    "asd",
						MType: "gauge",
						Value: &float1,
					})),
			},
			wantJSON{
				http.StatusOK, "application/json",
				"{\"id\":\"asd\",\"type\":\"gauge\",\"value\":123}"},
		},
		{
			"gauge post one more time case",
			reqJSON{
				http.MethodPost,
				"/update/",
				string(networkmsg.CreatePostUpdateMessage(
					networkmsg.Metrics{
						ID:    "asd",
						MType: "gauge",
						Value: &float1,
					})),
			},
			wantJSON{
				http.StatusOK, "application/json",
				"{\"id\":\"asd\",\"type\":\"gauge\",\"value\":123}"},
		},
		{
			"gauge post one more time case",
			reqJSON{
				http.MethodPost,
				"/update/",
				string(networkmsg.CreatePostUpdateMessage(
					networkmsg.Metrics{
						ID:    "asdf",
						MType: "gauge",
						Value: &float2,
					})),
			},
			wantJSON{
				http.StatusOK, "application/json",
				"{\"id\":\"asdf\",\"type\":\"gauge\",\"value\":123.23}"},
		},
		{
			"gauge get increased valid case",
			reqJSON{
				http.MethodPost,
				"/value/",
				string(networkmsg.CreatePostUpdateMessage(
					networkmsg.Metrics{
						ID:    "asd",
						MType: "gauge",
						Value: &float1,
					})),
			},
			wantJSON{
				http.StatusOK, "application/json",
				"{\"id\":\"asd\",\"type\":\"gauge\",\"value\":123}"},
		},
		{
			"gauge get missing value",
			reqJSON{
				http.MethodPost,
				"/value/",
				string(networkmsg.CreatePostUpdateMessage(
					networkmsg.Metrics{
						ID:    "asds",
						MType: "gauge",
					})),
			},
			wantJSON{
				http.StatusNotFound, "text/plain; charset=utf-8",
				"invalid counter name\n"},
		},
	}
	runJSONTests(t, &gaugeTests, ts)
}

func runJSONTests(t *testing.T, tests *[]testJSON, ts *httptest.Server) {
	for _, tt := range *tests {
		t.Run(tt.name, func(t *testing.T) {
			body := bytes.NewBufferString(tt.req.body)
			req, errReq := http.NewRequest(tt.req.method, ts.URL+tt.req.url, body)
			require.NoError(t, errReq)

			req.Header.Add("Content-Type", "application/json")

			resp, errResp := ts.Client().Do(req)
			assert.NoError(t, errResp)
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			//assert.Equal(t, tt.req.method, )
			assert.Equal(t, tt.want.body, string(respBody))
			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
		})
	}
}

// DEFAULT HANDLERS.
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

func TestBadRequestHandlerBaseController(t *testing.T) {
	cfg := config.Config{
		Host:     "localhost:8080",
		LogLevel: "Info",
	}

	log, err := logger.CreateZapLogger(cfg.LogLevel)
	if err != nil {
		panic(err)
	}
	//defer log.Sync()
	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(storageManager)

	ts := httptest.NewServer(CreateBase(cfg, log, storage).Route())
	defer ts.Close()

	badRequestTests := []test{
		//badRequestHandler
		{
			"post invalid path",
			req{http.MethodPost, "/update/count/fg/dfgdfg/dfg"},
			want{http.StatusBadRequest, "", ""},
		},
		{
			"get invalid path",
			req{http.MethodGet, "/update/count/sdfgdf/dfgdfg/gg"},
			want{http.StatusBadRequest, "", ""},
		},
		{
			"get invalid path",
			req{http.MethodGet, "/sdf/"},
			want{http.StatusMethodNotAllowed, "", ""},
		},
	}
	runTests(t, &badRequestTests, ts)
}

func TestListHandlerBaseController(t *testing.T) {
	cfg := config.Config{
		Host:     "localhost:8080",
		LogLevel: "Info",
	}

	log, err := logger.CreateZapLogger(cfg.LogLevel)
	if err != nil {
		panic(err)
	}
	//defer log.Sync()
	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(storageManager)

	ts := httptest.NewServer(CreateBase(cfg, log, storage).Route())
	defer ts.Close()

	badRequestTests := []test{
		//badRequestHandler
		{
			"list of params valid",
			req{http.MethodGet, "/"},
			want{http.StatusOK, "\n<html><body>\n<caption>GAUGES</caption>\n<table border = 2>\n</table>\n\n<caption>COUNTERS</caption>\n<table border = 2>\n</table>\n</body></html>\n", "text/html; charset=utf-8"},
		},
	}
	runTests(t, &badRequestTests, ts)
}

func TestMissingNameBaseController(t *testing.T) {
	cfg := config.Config{
		Host:     "localhost:8080",
		LogLevel: "Info",
	}

	log, err := logger.CreateZapLogger(cfg.LogLevel)
	if err != nil {
		panic(err)
	}
	//defer log.Sync()
	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(storageManager)

	ts := httptest.NewServer(CreateBase(cfg, log, storage).Route())
	defer ts.Close()
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
}

func TestCounterBaseController(t *testing.T) {
	cfg := config.Config{
		Host:     "localhost:8080",
		LogLevel: "Info",
	}

	log, err := logger.CreateZapLogger(cfg.LogLevel)
	if err != nil {
		panic(err)
	}
	//defer log.Sync()
	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(storageManager)

	ts := httptest.NewServer(CreateBase(cfg, log, storage).Route())
	defer ts.Close()

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
}

func TestGaugeBaseController(t *testing.T) {
	cfg := config.Config{
		Host:     "localhost:8080",
		LogLevel: "Info",
	}

	log, err := logger.CreateZapLogger(cfg.LogLevel)
	if err != nil {
		panic(err)
	}
	//defer log.Sync()
	storageManager := storagemngr.CreateFileManager(cfg.StoragePath, log)
	storage := memstorage.Create(storageManager)

	ts := httptest.NewServer(CreateBase(cfg, log, storage).Route())
	defer ts.Close()

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

func runTests(t *testing.T, tests *[]test, ts *httptest.Server) {
	for _, tt := range *tests {
		t.Run(tt.name, func(t *testing.T) {
			req, errReq := http.NewRequest(tt.req.method, ts.URL+tt.req.url, nil)
			require.NoError(t, errReq)

			req.Header.Add("Content-Type", "html/text")
			req.Header.Add("Accept-Encoding", "gzip")

			resp, errResp := ts.Client().Do(req)
			assert.NoError(t, errResp)
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			response, _ := compressor.GzipDecompress(respBody)

			assert.Equal(t, tt.want.response, string(response))
			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
		})
	}
}

func TestBaseController_checkStorageHandler(t *testing.T) {
	type fields struct {
		config     config.Config
		storage    memstorage.MemStorage
		logger     logger.BaseLogger
		compressor compressor.GzipHandler
	}
	type args struct {
		w   http.ResponseWriter
		in1 *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &BaseController{
				config:     tt.fields.config,
				storage:    tt.fields.storage,
				logger:     tt.fields.logger,
				compressor: tt.fields.compressor,
			}
			c.checkStorageHandler(tt.args.w, tt.args.in1)
		})
	}
}
