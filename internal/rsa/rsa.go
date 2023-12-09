package rsa

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Encoder RSA message encryptor.
type Encoder struct {
	key *rsa.PublicKey
}

// CreateEncoder creates RSA encoder from cert file.
func CreateEncoder(certFilePath string) (*Encoder, error) {
	certPEM, err := os.ReadFile(certFilePath)
	if err != nil {
		return nil, fmt.Errorf("read RSA cert: %w", err)
	}

	certBlock, _ := pem.Decode(certPEM)
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse RSA public key: %w", err)
	}

	return &Encoder{
		key: cert.PublicKey.(*rsa.PublicKey),
	}, nil
}

// Encode encrypts message using RSA public key.
func (e *Encoder) Encode(msg []byte) ([]byte, error) {
	if e.key == nil {
		return nil, fmt.Errorf("RSA cert is not set")
	}

	return rsa.EncryptPKCS1v15(rand.Reader, e.key, msg)
}

// Decoder RSA message decryptor.
type Decoder struct {
	key *rsa.PrivateKey
}

// CreateDecoder creates RSA encoder from cert file.
func CreateDecoder(keyFilePath string) (*Decoder, error) {
	keyPEM, err := os.ReadFile(keyFilePath)
	if err != nil {
		return nil, fmt.Errorf("read RSA key: %w", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse RSA public key: %w", err)
	}

	return &Decoder{
		key: key,
	}, nil
}

// Decode encrypts message using RSA public key.
func (e *Decoder) Decode(msg []byte) ([]byte, error) {
	if e.key == nil {
		return nil, fmt.Errorf("RSA cert is not set")
	}

	return rsa.DecryptPKCS1v15(rand.Reader, e.key, msg)
}

// DecodeRSAHandler handler for body decoding using RSA private key.
func (e *Decoder) DecodeRSAHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() {
			_ = r.Body.Close()
		}()

		decryptedMessage, err := e.Decode(buf.Bytes())
		if err != nil {
			http.Error(w, "Error decrypting message", http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(decryptedMessage))
		next.ServeHTTP(w, r)
	})
}
