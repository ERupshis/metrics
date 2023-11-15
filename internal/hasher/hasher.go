// Package hasher provides hasher for message hash-sum calculation and verification.
package hasher

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"net/http"

	"github.com/erupshis/metrics/internal/logger"
)

const (
	SHA256 = iota // type of used hash algorithm
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

// Hasher stores hash related config data.
type Hasher struct {
	log      logger.BaseLogger
	hashType int    // type of algorithm
	key      string // hash key
}

// CreateHasher create method.
func CreateHasher(hashKey string, hashType int, log logger.BaseLogger) *Hasher {
	return &Hasher{key: hashKey, hashType: hashType, log: log}
}

// Handler middleware handler.
// Validates incoming messages and check hash-sum.
func (hr *Hasher) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hashHeaderValue := r.Header.Get(hr.GetHeader())
		if hashHeaderValue != "" {
			var buf bytes.Buffer
			_, err := io.Copy(&buf, r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			ok, err := hr.isRequestValid(hashHeaderValue, buf)
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

// WriteHashHeaderInResponseIfNeed calculates hash for responseBody if hashKey was assigned.
func (hr *Hasher) WriteHashHeaderInResponseIfNeed(w http.ResponseWriter, responseBody []byte) {
	if hr.key == "" {
		return
	}

	hashValue, err := hr.HashMsg(responseBody)
	if err != nil {
		hr.log.Info("[Hasher::WriteHashHeaderInResponseIfNeed] failed to add hasher in response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add(hr.GetHeader(), hashValue)
}

// HashMsg returns hash for message.
func (hr *Hasher) HashMsg(msg []byte) (string, error) {
	switch hr.getAlgo() {
	case algoSHA256:
		return hashMsg(sha256.New, msg, hr.key)
	default:
		panic("unknown algorithm")
	}
}

// hashMsg returns hash for message.
func hashMsg(hashFunc func() hash.Hash, msg []byte, key string) (string, error) {
	var h hash.Hash
	if key != "" {
		h = hmac.New(hashFunc, []byte(key))
	} else {
		// hasher sum w/o authentication.
		h = hashFunc()
	}

	_, err := h.Write(msg)
	if err != nil {
		return "", err
	}

	hashVal := h.Sum(nil)
	return fmt.Sprintf("%x", hashVal), nil
}

// isRequestValid validates incoming message and compare incoming and calculated hashes.
func (hr *Hasher) isRequestValid(hashHeaderValue string, buffer bytes.Buffer) (bool, error) {
	ok, err := hr.checkRequestHash(hashHeaderValue, buffer.Bytes())
	if err != nil {
		return false, fmt.Errorf("hasher validation: %w", err)
	}

	return ok, nil
}

// checkRequestHash validates incoming message and compare incoming and calculated hashes.
func (hr *Hasher) checkRequestHash(hashHeaderValue string, body []byte) (bool, error) {
	if hr.key == "" {
		return true, nil
	}

	if hashHeaderValue == "" {
		return true, nil
	}

	hashValue, err := hr.HashMsg(body)
	if err != nil {
		return false, fmt.Errorf("check request hasher with SHA256: %w", err)
	}

	return hashHeaderValue == hashValue, nil
}

// GetHeader returns http Header key of used hash type.
func (hr *Hasher) GetHeader() string {
	switch hr.hashType {
	case SHA256:
		return headerSHA256
	default:
		return headerSHA256
	}
}

// getAlgo returns used algo.
func (hr *Hasher) getAlgo() int {
	switch hr.hashType {
	case SHA256:
		return algoSHA256
	default:
		return algoSHA256
	}
}

// GetKey returns hash key.
func (hr *Hasher) GetKey() string {
	return hr.key
}
