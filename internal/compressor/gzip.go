package compressor

import (
	"compress/gzip"
	"io"
	"net/http"
)

// compressWriter http.ResponseWriter's decorator.
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// newGzipCompressWriter create method.
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

// Reset overwrites base inner http.ResponseWriter entity.
func (c *compressWriter) Reset(w http.ResponseWriter) {
	if c.zw == nil {
		return
	}
	c.w = w
	c.zw.Reset(c.w)
}

// Header returns inner http.ResponseWriter header.
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write writes compressed data in inner http.ResponseWriter body.
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader writes header in inner http.ResponseWriter header.
// If statusCode < 300 - adds Content-Encoding = gzip.
func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close closes gzip.Writer with flushing any unwritten data + gzip footer.
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// compressWriter io.ReadCloser's decorator.
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// newGzipCompressWriter create method.
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

// Reset overwrites base inner io.ReadCloser entity.
func (c *compressReader) Reset(r io.ReadCloser) error {
	if c.zr == nil {
		return nil
	}
	c.r = r
	return c.zr.Reset(r)
}

// Read reads uncompressed data from inner io.ReadCloser.
func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close closes reader.
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
