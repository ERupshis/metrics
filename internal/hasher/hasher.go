package hasher

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"github.com/erupshis/metrics/internal/logger"
	"hash"
	"io"
	"net/http"
)

const (
	SHA256 = iota
)

const (
	headerSHA256 = "HashSHA256"
)

const (
	algoSHA256 = iota
)

type readCloserWrapper struct {
	io.Reader
	io.Closer
}

type Hasher struct {
	log      logger.BaseLogger
	hashType int
}

func CreateHasher(hashType int, log logger.BaseLogger) *Hasher {
	return &Hasher{hashType: hashType, log: log}
}

func (hr *Hasher) Handler(h http.Handler, hashKey string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hashHeaderValue := r.Header.Get(hr.GetHeader())
		if hashHeaderValue != "" {
			var buf bytes.Buffer
			_, err := io.Copy(&buf, r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			ok, err := hr.isRequestValid(hashHeaderValue, hashKey, buf)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			rc := &readCloserWrapper{
				Reader: bytes.NewReader(buf.Bytes()),
				Closer: r.Body,
			}
			r.Body = rc
		}

		h.ServeHTTP(w, r)
	})
}

func (hr *Hasher) WriteHashHeaderInResponseIfNeed(w http.ResponseWriter, hashKey string, responseBody []byte) error {
	if hashKey == "" {
		return nil
	}

	hashValue, err := hr.HashMsg(responseBody, hashKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	w.Header().Add(hr.GetHeader(), hashValue)
	return nil
}

func (hr *Hasher) HashMsg(msg []byte, key string) (string, error) {
	switch hr.getAlgo() {
	case algoSHA256:
		return hashMsg(sha256.New, msg, key)
	default:
		panic("unknown algorithm")
	}
}

func hashMsg(hashFunc func() hash.Hash, msg []byte, key string) (string, error) {
	var h hash.Hash
	if key != "" {
		h = hmac.New(hashFunc, []byte(key))
	} else {
		//hash sum w/o authentication.
		h = hashFunc()
	}

	_, err := h.Write(msg)
	if err != nil {
		return "", err
	}

	hashVal := h.Sum(nil)
	return fmt.Sprintf("%x", hashVal), nil
}

func (hr *Hasher) isRequestValid(hashHeaderValue string, hashKey string, buffer bytes.Buffer) (bool, error) {
	ok, err := hr.checkRequestHash(hashHeaderValue, hashKey, buffer.Bytes())
	if err != nil {
		return false, fmt.Errorf("hash validation: %w", err)
	}

	return ok, nil
}

func (hr *Hasher) checkRequestHash(hashHeaderValue string, hashKey string, body []byte) (bool, error) {
	if hashKey == "" {
		return true, nil
	}

	if hashHeaderValue == "" {
		return true, nil
	}

	hashValue, err := hr.HashMsg(body, hashKey)
	if err != nil {
		return false, fmt.Errorf("check request hash with SHA256: %w", err)
	}

	return hashHeaderValue == hashValue, nil
}

func (hr *Hasher) GetHeader() string {
	switch hr.hashType {
	case SHA256:
		return headerSHA256
	default:
		return headerSHA256
	}
}

func (hr *Hasher) getAlgo() int {
	switch hr.hashType {
	case SHA256:
		return algoSHA256
	default:
		return algoSHA256
	}
}
