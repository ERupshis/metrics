package hasher

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"hash"
	"net/http"
)

const (
	HeaderSHA256 = "HashSHA256"
)

const (
	AlgoSHA256 = iota
)

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

func CheckRequestHash(r *http.Request, hashHeaderKey string, hashKey string, body []byte) (bool, error) {
	if hashKey == "" {
		return true, nil
	}

	reqHashValue := r.Header.Get(hashHeaderKey)
	if reqHashValue == "" {
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

	return reqHashValue == hashValue, nil
}
