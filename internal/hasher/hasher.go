package hasher

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"net/http"
)

const (
	HeaderSHA256 = "HashSHA256"
)

const (
	AlgoSHA256 = iota
)

type ReadCloserWrapper struct {
	io.Reader
	io.Closer
}

func Handler(h http.Handler, hashKey string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hashHeaderValue := r.Header.Get(HeaderSHA256)
		if hashHeaderValue != "" {
			var buf bytes.Buffer
			_, err := io.Copy(&buf, r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			ok, err := isRequestValid(HeaderSHA256, hashHeaderValue, hashKey, buf)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			rc := &ReadCloserWrapper{
				Reader: bytes.NewReader(buf.Bytes()),
				Closer: r.Body,
			}
			r.Body = rc
		}

		h.ServeHTTP(w, r)
	})
}

func HashMsg(algo uint, msg []byte, key string) (string, error) {
	switch algo {
	case AlgoSHA256:
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

func isRequestValid(hashHeaderKey string, hashHeaderValue string, hashKey string, buffer bytes.Buffer) (bool, error) {
	ok, err := CheckRequestHash(hashHeaderKey, hashHeaderValue, hashKey, buffer.Bytes())
	if err != nil {
		return false, fmt.Errorf("hash validation: %w", err)
	}

	return ok, nil
}

func CheckRequestHash(hashHeaderKey string, hashHeaderValue string, hashKey string, body []byte) (bool, error) {
	if hashKey == "" {
		return true, nil
	}

	if hashHeaderValue == "" {
		return true, nil
	}

	var hashValue string
	var err error
	switch hashHeaderKey {
	case HeaderSHA256:
		hashValue, err = HashMsg(AlgoSHA256, body, hashKey)
		if err != nil {
			return false, fmt.Errorf("check request hash with SHA256: %w", err)
		}
	default:
		panic("unknown header key")
	}

	return hashHeaderValue == hashValue, nil
}
