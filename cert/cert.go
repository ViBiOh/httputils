package cert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"
)

func pemBlockForKey(ecdsaKey *ecdsa.PrivateKey) (*pem.Block, error) {
	key, err := x509.MarshalECPrivateKey(ecdsaKey)
	if err != nil {
		return nil, fmt.Errorf(`Error while marshalling private key: %v`, err)
	}
	return &pem.Block{Type: "EC PRIVATE KEY", Bytes: key}, nil
}

func getLocalIps() ([]net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf(`Error while getting interface addrs: %v`, err)
	}

	ips := make([]net.IP, 0)

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP)
			}
		}
	}
	return ips, nil
}

// GenerateCert self signed with CA for use with TLS
func GenerateCert(organization string, hosts []string) error {
	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf(`Error while generating cert key: %v`, err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf(`Error while generating serial number: %v`, err)
	}

	startDate := time.Now()
	endDate := startDate.Add(time.Hour * 24 * 265)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{organization},
		},
		NotBefore: startDate,
		NotAfter:  endDate,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA: true,
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	ips, err := getLocalIps()
	if err != nil {
		return fmt.Errorf(`Error while getting locals ips: %v`, err)
	}

	for _, ip := range ips {
		template.IPAddresses = append(template.IPAddresses, ip)
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &ecdsaKey.PublicKey, ecdsaKey)
	if err != nil {
		return fmt.Errorf(`Error while creating certificate: %v`, err)
	}

	certFile, err := os.Create(`cert.pem`)
	if err != nil {
		return fmt.Errorf(`Error while creating certificate file: %v`, err)
	}
	defer certFile.Close()

	pem.Encode(certFile, &pem.Block{Type: `CERTIFICATE`, Bytes: der})

	keyFile, err := os.OpenFile(`key.pem`, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf(`Error while creating key file: %v`, err)
	}
	defer keyFile.Close()

	pemBlock, err := pemBlockForKey(ecdsaKey)
	if err != nil {
		return fmt.Errorf(`Error while getting pem block for key: %v`, err)
	}

	pem.Encode(keyFile, pemBlock)

	return nil
}
