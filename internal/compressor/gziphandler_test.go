package compressor

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGzipCompressDecompress(t *testing.T) {
	tests := []struct {
		name string
		args []byte
		want []byte
	}{
		{
			name: "valid case",
			args: []byte(`{"should":"be fine, "should":"be fine", "should":"be fine", "should":"be fine", "should":"be fine"}"`),
			want: []byte(`{"should":"be fine, "should":"be fine", "should":"be fine", "should":"be fine", "should":"be fine"}"`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressed, err := GzipCompress(tt.args)
			if err != nil || string(compressed) == string(tt.args) {
				t.Errorf("GzipCompress() finished with error = %v or input equal to output", err)
				return
			}
			if len(compressed) > len(tt.args) {
				t.Errorf("GzipCompress() compressed len > source len")
			}

			decompressed, err := GzipDecompress(compressed)
			if err != nil || string(compressed) == string(decompressed) {
				t.Errorf("GzipCompress() finished with error = %v or input equal to output", err)
			}
			if !reflect.DeepEqual(decompressed, tt.want) {
				t.Errorf("GzipDecompress() got = %v, want %v", decompressed, tt.want)
			}
		})
	}
}

func Test_canCompress(t *testing.T) {
	type args struct {
		accept      string
		contentType string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid positive Accept(JSON)",
			args: args{accept: "application/json"},
			want: true,
		},
		{
			name: "valid positive Accept(HTML)",
			args: args{accept: "html/text"},
			want: true,
		},
		{
			name: "valid negative Accept",
			args: args{accept: "wrong"},
			want: false,
		},
		{
			name: "valid positive Content-Type(JSON)",
			args: args{contentType: "application/json"},
			want: true,
		},
		{
			name: "valid positive Content-Type(HTML)",
			args: args{contentType: "html/text"},
			want: true,
		},
		{
			name: "valid negative Accept",
			args: args{contentType: "wrong"},
			want: false,
		},
	}
	for _, ttCommon := range tests {
		tt := ttCommon
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			buf := bytes.NewBufferString("")
			req := httptest.NewRequest("POST", "/", buf)
			req.RequestURI = ""
			if tt.args.accept != "" {
				req.Header.Set("Accept", tt.args.accept)
			}
			if tt.args.contentType != "" {
				req.Header.Set("Content-Type", tt.args.contentType)
			}

			if got := canCompress(req); got != tt.want {
				t.Errorf("canCompress() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ...

func TestGzipHandler(t *testing.T) {
	var in = `{"REQUEST":"IN"}`
	var out = `{"REQUEST":"OUT"}`

	compressor := GzipHandler{}

	handler := func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer func() {
			_ = r.Body.Close()
		}()

		if buf.String() != in {
			http.Error(w, "decompressed input is incorrect.", http.StatusBadRequest)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write([]byte(out))
	}
	gzHandler := compressor.GzipHandle(http.HandlerFunc(handler))

	srv := httptest.NewServer(gzHandler)
	defer srv.Close()
	requestBody := in

	// ожидаемое содержимое тела ответа при успешном запросе
	successBody := out

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest(http.MethodGet, srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		defer func() {
			_ = resp.Body.Close()
		}()

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.JSONEq(t, successBody, string(b))
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)
		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Accept-Encoding", "gzip")
		r.Header.Set("Accept", "application/json")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		defer func() {
			_ = resp.Body.Close()
		}()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.JSONEq(t, successBody, string(b))
	})
}
