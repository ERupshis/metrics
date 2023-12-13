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

type expiresAfter struct {
	years  int
	months int
	days   int
}

var (
	validTo      = expiresAfter{days: 0, months: 0, years: 10}
	subjectKeyID = []byte{1, 2, 3, 4, 6}
	ipAddress    = net.IPv4(127, 0, 0, 1)

	pathToPrivateKey = "rsa/key.pem"
	pathToCert       = "rsa/cert.pem"

	organization = []string{"erupshis.metrics"}
	country      = []string{"RU"}
)

func main() {
	// creates template of certificate.
	certTemplate := &x509.Certificate{
		// unique certificate number.
		SerialNumber: big.NewInt(436654756),
		// base information about certificate owner.
		Subject: pkix.Name{
			Organization: organization,
			Country:      country,
		},

		IPAddresses: []net.IP{ipAddress, net.IPv6loopback},
		// certificate is valid from NOW.
		NotBefore: time.Now(),
		// lifetime: 10 years
		NotAfter:     time.Now().AddDate(validTo.years, validTo.months, validTo.days),
		SubjectKeyId: subjectKeyID,
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

	if err = os.WriteFile(pathToCert, certPEM, 0644); err != nil {
		log.Fatal(err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	if privateKeyPEM == nil {
		log.Fatal("Failed to encode key to PEM")
	}
	if err = os.WriteFile(pathToPrivateKey, privateKeyPEM, 0600); err != nil {
		log.Fatal(err)
	}

}
