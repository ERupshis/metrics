package handlers

import (
	"github.com/ERupshis/metrics/internal/server/memstorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type request struct {
	reqType string
	path    string
}

type want struct {
	code        int
	response    string
	contentType string
}

func TestInvalid(t *testing.T) {
	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			"simple test",
			request{http.MethodGet, "/asd"},
			want{http.StatusBadRequest, ``, ""},
		},
		{
			"simple test",
			request{http.MethodGet, "/"},
			want{http.StatusBadRequest, ``, ""},
		},
	}

	handler := Create(memstorage.Create())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.request.reqType, tt.request.path, nil)
			writer := httptest.NewRecorder()
			handler.Invalid(writer, request)
			res := writer.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, []byte(tt.want.response), resBody)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestMissingName(t *testing.T) {
	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			"simple test",
			request{http.MethodGet, "/asd"},
			want{http.StatusNotFound, ``, ""},
		},
		{
			"simple test",
			request{http.MethodGet, "/"},
			want{http.StatusNotFound, ``, ""},
		},
	}

	handler := Create(memstorage.Create())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.request.reqType, tt.request.path, nil)
			writer := httptest.NewRecorder()
			handler.MissingName(writer, request)
			res := writer.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, []byte(tt.want.response), resBody)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
