package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

func main() {
	// creates template of certificate.
	certTemplate := &x509.Certificate{
		// unique certificate number.
		SerialNumber: big.NewInt(436654756),
		// base information about certificate owner.
		Subject: pkix.Name{
			Organization: []string{"erupshis.metrics"},
			Country:      []string{"RU"},
		},

		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		// certificate is valid from NOW.
		NotBefore: time.Now(),
		// lifetime: 10 years
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		// install usage of key for digital signature, and client and server authorization.
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	// generating new private RSA-key with length 4096 bytes
	// make attention. for generation certificate and key
	// rand.Reader is used as source of random numbers.
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatal(err)
	}

	// generate certificate x.509
	certBytes, err := x509.CreateCertificate(rand.Reader, certTemplate, certTemplate, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Fatal(err)
	}

	// encoding certificate and key in PEM format, which is used for storing and exchange by crypto keys.
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	if certPEM == nil {
		log.Fatal("Failed to encode certificate to PEM")
	}

	if err = os.WriteFile("rsa/cert.pem", certPEM, 0644); err != nil {
		log.Fatal(err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	if privateKeyPEM == nil {
		log.Fatal("Failed to encode key to PEM")
	}
	if err = os.WriteFile("rsa/key.pem", privateKeyPEM, 0600); err != nil {
		log.Fatal(err)
	}

}
