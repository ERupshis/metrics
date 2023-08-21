package compressor

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"strings"
)

var availableContentTypes = []string{"application/json", "html/text", "text/html"}

type GzipHandler struct {
	writer *compressWriter
	reader *compressReader
}

func (gz *GzipHandler) GzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip && canCompress(r) {
			if gz.writer == nil {
				gz.writer = newGzipCompressWriter(w)
			} else {
				gz.writer.Reset(w)
			}
			ow = gz.writer

			w.Header().Set("Content-Encoding", "gzip")
			defer gz.writer.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			if gz.reader == nil {
				var err error
				gz.reader, err = newGzipCompressReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			} else {
				gz.reader.Reset(r.Body)
			}

			r.Body = gz.reader
			defer gz.reader.Close()
		}

		next.ServeHTTP(ow, r)
	})
}

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

func GzipCompress(data []byte) ([]byte, error) {
	var b bytes.Buffer

	w, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("failed init compress writer: %v", err)
	}

	_, err = w.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %v", err)
	}

	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %v", err)
	}

	return b.Bytes(), nil
}

func GzipDecompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed init decompress reader: %v", err)
	}
	defer r.Close()

	var b bytes.Buffer
	_, err = b.ReadFrom(r)
	if err != nil {
		return nil, fmt.Errorf("failed decompress data: %v", err)
	}

	return b.Bytes(), nil
}
