package hasher

import (
	"crypto/hmac"
	"crypto/sha256"
	"hash"
)

const (
	SHA256 = iota
)

func HashMsg(algo uint, msg []byte, key string) (string, error) {
	switch algo {
	case SHA256:
		return hashMsgSHA256(msg, key)
	default:
		panic("unknown algorithm")
	}
}

func hashMsgSHA256(msg []byte, key string) (string, error) {
	var h hash.Hash
	if key != "" {
		h = hmac.New(sha256.New, []byte(key))
	} else {
		//hashing w/o authentication.
		h = sha256.New()
	}

	_, err := h.Write(msg)
	if err != nil {
		return "", err
	}

	hashVal := h.Sum(nil)
	return string(hashVal), nil
}
