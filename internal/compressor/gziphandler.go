package compressor

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// availableContentTypes types allowed to handle by compressor
var availableContentTypes = []string{"application/json", "html/text"}

// GzipHandler stores compressWriter & compressReader entities.
type GzipHandler struct {
	writer *compressWriter
	wrOnce sync.Once

	reader *compressReader
	rdmux  sync.Mutex
}

// setGzipCompWriter assigns compress writer.
// Creates new one if it was not created before, otherwise Reset existing by new responseWriter.
func (gz *GzipHandler) setGzipCompWriter(w http.ResponseWriter) {
	gz.wrOnce.Do(func() {
		gz.writer = newGzipCompressWriter(w)
	})

	gz.writer.Reset(w)
}

// setGzipCompReader assigns compress reader.
// Creates new one if it was not created before, otherwise Reset existing by new responseReader.
func (gz *GzipHandler) setGzipCompReader(r *http.Request) error {
	gz.rdmux.Lock()
	defer gz.rdmux.Unlock()

	if gz.reader == nil {
		var err error
		gz.reader, err = newGzipCompressReader(r.Body)
		if err != nil {
			return fmt.Errorf("set gzip reader: %w", err)
		}
	} else {
		return gz.reader.Reset(r.Body)
	}

	return nil
}

// GzipHandle middleware handler.
// Performs compression response and decompression request if Content-Type corresponds to availableContentTypes
// and Accept-Encoding allows to use gzip compression.
func (gz *GzipHandler) GzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip && canCompress(r) {
			gz.setGzipCompWriter(w)
			ow = gz.writer

			w.Header().Set("Content-Encoding", "gzip")
			defer func() {
				_ = gz.writer.Close()
			}()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			if err := gz.setGzipCompReader(r); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Body = gz.reader
			defer func() {
				_ = gz.reader.Close()
			}()
		}

		next.ServeHTTP(ow, r)
	})
}

// canCompress support method. Checks if Content-Typ value allows to use compressor.
func canCompress(req *http.Request) bool {
	for _, contType := range availableContentTypes {
		for _, value := range req.Header.Values("Accept") {
			if strings.Contains(value, contType) {
				return true
			}
		}
		for _, value := range req.Header.Values("Content-Type") {
			if strings.Contains(value, contType) {
				return true
			}
		}
	}

	return false
}

// GzipCompress provides data's gzip compression.
func GzipCompress(data []byte) ([]byte, error) {
	var b bytes.Buffer

	w, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("failed init compress writer: %w", err)
	}

	_, err = w.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %w", err)
	}

	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %w", err)
	}

	return b.Bytes(), nil
}

// GzipDecompress provides data's gzip decompression.
func GzipDecompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed init decompress reader: %w", err)
	}
	defer func() {
		_ = r.Close()
	}()

	var b bytes.Buffer
	_, err = b.ReadFrom(r)
	if err != nil {
		return nil, fmt.Errorf("failed decompress data: %w", err)
	}

	return b.Bytes(), nil
}
