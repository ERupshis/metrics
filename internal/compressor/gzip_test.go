package compressor

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompressWriter(t *testing.T) {
	responseRecorder := httptest.NewRecorder()
	cw := newGzipCompressWriter(responseRecorder)

	data := []byte("test data")
	if _, err := cw.Write(data); err != nil {
		t.Errorf("failed to write data: %v", err)
	}

	if err := cw.Close(); err != nil {
		t.Errorf("failed to close writer")
	}

	compressedData, err := gzip.NewReader(bytes.NewReader(responseRecorder.Body.Bytes()))
	if err != nil {
		t.Fatal(err)
	}

	buf := bytes.Buffer{}
	if _, err = buf.ReadFrom(compressedData); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(data, buf.Bytes()) {
		t.Errorf("Expected compressed data to be %v, got %v", data, buf.Bytes())
	}

	cw.WriteHeader(http.StatusOK)
	if responseRecorder.Header().Get("Content-Encoding") != "gzip" {
		t.Error("Expected Content-Encoding header to be 'gzip'")
	}

	if err = cw.Close(); err != nil {
		t.Errorf("failed to close writer")
	}

	if err := compressedData.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestCompressReader(t *testing.T) {
	compressedData := compressString("test data")
	fakeCloserEntity := &fakeCloser{Reader: bytes.NewReader(compressedData)}

	cr, err := newGzipCompressReader(fakeCloserEntity)
	if err != nil {
		t.Fatal(err)
	}

	buf := bytes.Buffer{}
	if _, err = buf.ReadFrom(cr); err != nil {
		t.Fatal(err)
	}

	expectedData := []byte("test data")
	if !bytes.Equal(buf.Bytes(), expectedData) {
		t.Errorf("Expected read data to be %v, got %v", expectedData, buf.Bytes())
	}

	if err = cr.Close(); err != nil {
		t.Fatal(err)
	}

	if !fakeCloserEntity.closed {
		t.Error("Expected fakeCloser to be closed")
	}
}

// Helper function to compress a string using gzip
func compressString(s string) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, err := w.Write([]byte(s))
	if err != nil {
		panic(err)
	}
	if err := w.Close(); err != nil {
		panic(err)
	}
	return b.Bytes()
}

// fakeCloser реализует интерфейс io.ReadCloser для использования в тестах
type fakeCloser struct {
	io.Reader
	closed bool
}

func (f *fakeCloser) Close() error {
	f.closed = true
	return nil
}

func TestCompressWriterReset(t *testing.T) {
	// Create a fake http.ResponseWriter
	responseRecorder := httptest.NewRecorder()

	// Create a compressWriter
	cw := newGzipCompressWriter(responseRecorder)
	if cw == nil {
		t.Fatal("Failed to create compressWriter")
	}

	// Write some data
	data := []byte("test data")
	if _, err := cw.Write(data); err != nil {
		t.Errorf("failed to write data: %v", err)
	}

	// Call Reset with a new fake http.ResponseWriter
	newResponseRecorder := httptest.NewRecorder()
	cw.Reset(newResponseRecorder)

	// Verify that the compressWriter is reset
	if cw.w != newResponseRecorder {
		t.Error("Reset did not update the http.ResponseWriter")
	}

	// Verify that the compressWriter's internal gzip.Writer is reset
	if cw.zw == nil {
		t.Error("Reset did not recreate the internal gzip.Writer")
	}

	// Verify that the header is cleared after calling Reset
	if len(cw.w.Header()) > 0 {
		t.Error("Reset did not clear the headers")
	}
}

func TestCompressReaderReset(t *testing.T) {
	// Create a fake io.ReadCloser
	compressedData := compressString("test data")
	fakeCloserEntity := &fakeCloser{Reader: bytes.NewReader(compressedData)}

	// Create a compressReader
	cr, err := newGzipCompressReader(fakeCloserEntity)
	if err != nil {
		t.Fatal(err)
	}

	// Read some data
	buf := bytes.Buffer{}
	if _, err = buf.ReadFrom(cr); err != nil {
		t.Fatalf("Error reading data: %v", err)
	}

	expectedData := []byte("test data")
	if !bytes.Equal(buf.Bytes(), expectedData) {
		t.Errorf("Expected read data to be %v, got %v", expectedData, buf.Bytes())
	}

	// Call Reset with a new fake io.ReadCloser
	newCompressedData := compressString("new test data")
	newFakeCloserEntity := &fakeCloser{Reader: bytes.NewReader(newCompressedData)}
	err = cr.Reset(newFakeCloserEntity)
	require.NoError(t, err)

	// Read data again
	newBuf := bytes.Buffer{}
	if _, err = newBuf.ReadFrom(cr); err != nil {
		t.Fatalf("Error reading data: %v", err)
	}

	// Verify that the compressReader is reset and can be reused
	if !bytes.Equal(newBuf.Bytes(), []byte("new test data")) {
		t.Errorf("Expected read data after Reset to be 'new test data', got %v", newBuf.Bytes())
	}

	// Verify that the compressReader's internal gzip.Reader is reset
	if cr.zr == nil {
		t.Error("Reset did not recreate the internal gzip.Reader")
	}
}
