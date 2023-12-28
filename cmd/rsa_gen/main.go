package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
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
	validTo   = expiresAfter{days: 0, months: 0, years: 10}
	ipAddress = net.IPv4(127, 0, 0, 1)

	pathToPrivateKey = "rsa/key.pem"
	pathToCert       = "rsa/cert.pem"
	pathToCACert     = "rsa/ca_cert.pem"

	caOrganization = []string{"erupshis.metrics"}
	caCountry      = []string{"RU"}
)

var (
	// Для подписи запросов на сертификат (CSR)
	caPrivateKey, _ = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
)

func main() {
	caCertTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: caOrganization,
			Country:      caCountry,
		},
		IPAddresses:           []net.IP{ipAddress, net.IPv6loopback},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(validTo.years, validTo.months, validTo.days),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Создание сертификата для CA
	caCertBytes, err := x509.CreateCertificate(rand.Reader, caCertTemplate, caCertTemplate, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		log.Fatal(err)
	}

	// Сохранение сертификата CA в файл
	caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertBytes})
	if caCertPEM == nil {
		log.Fatal("Failed to encode CA certificate to PEM")
	}
	if err := os.WriteFile(pathToCACert, caCertPEM, 0644); err != nil {
		log.Fatal(err)
	}

	// Создание шаблона сертификата для сервера
	certTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization: caOrganization,
			Country:      caCountry,
		},
		IPAddresses: []net.IP{ipAddress, net.IPv6loopback},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(validTo.years, validTo.months, validTo.days),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
	}

	// Генерация закрытого ключа для сервера
	privateKey, err := rsa.GenerateKey(rand.Reader, 8192)
	if err != nil {
		log.Fatal(err)
	}

	// Генерация сертификата для сервера и подпись его CA
	certBytes, err := x509.CreateCertificate(rand.Reader, certTemplate, caCertTemplate, &privateKey.PublicKey, caPrivateKey)
	if err != nil {
		log.Fatal(err)
	}

	// Сохранение сертификата сервера в файл
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	if certPEM == nil {
		log.Fatal("Failed to encode certificate to PEM")
	}
	if err := os.WriteFile(pathToCert, certPEM, 0644); err != nil {
		log.Fatal(err)
	}

	// Сохранение закрытого ключа сервера в файл
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	if privateKeyPEM == nil {
		log.Fatal("Failed to encode key to PEM")
	}
	if err := os.WriteFile(pathToPrivateKey, privateKeyPEM, 0600); err != nil {
		log.Fatal(err)
	}

}
