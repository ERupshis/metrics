package compressor

import (
	"compress/gzip"
	"io"
	"net/http"
)

// WRITER

type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newGzipCompressWriter(w http.ResponseWriter) *compressWriter {
	gz, err := gzip.NewWriterLevel(w, gzip.BestCompression)
	if err != nil {
		return nil
	}

	return &compressWriter{
		w:  w,
		zw: gz,
	}
}

func (c *compressWriter) Reset(w http.ResponseWriter) {
	if c.zw == nil {
		return
	}
	c.w = w
	c.zw.Reset(c.w)
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// READER

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newGzipCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c *compressReader) Reset(r io.ReadCloser) error {
	if c.zr == nil {
		return nil
	}
	c.r = r
	return c.zr.Reset(r)
}

func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
