package cert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"time"
)

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
// from https://golang.org/src/net/http/server.go
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

func strSliceContains(slice []string, search string) bool {
	for _, value := range slice {
		if value == search {
			return true
		}
	}

	return false
}

// clneTLSConfig returns a shallow clone of cfg, or a new zero tls.Config if
// cfg is nil. This is safe to call even if cfg is in active use by a TLS
// client or server.
// from https://golang.org/src/net/http/transport.go
func cloneTLSConfig(cfg *tls.Config) *tls.Config {
	if cfg == nil {
		return &tls.Config{}
	}
	return cfg.Clone()
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

// ListenAndServeTLS with provided bytes of cert instead of file
// Largely inspired by https://golang.org/src/net/http/server.go
func ListenAndServeTLS(server *http.Server, certPEMBlock []byte, keyPEMBlock []byte) error {
	addr := server.Addr
	if addr == `` {
		addr = `:https`
	}

	config := cloneTLSConfig(server.TLSConfig)
	if !strSliceContains(config.NextProtos, "http/1.1") {
		config.NextProtos = append(config.NextProtos, "http/1.1")
	}

	config.Certificates = make([]tls.Certificate, 1)
	certificate, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return fmt.Errorf(`Error while getting x509 KeyPair: %v`, err)
	}
	config.Certificates[0] = certificate

	listener, err := net.Listen(`tcp`, addr)
	if err != nil {
		return fmt.Errorf(`Error while listening: %v`, err)
	}

	tlsListener := tls.NewListener(
		tcpKeepAliveListener{listener.(*net.TCPListener)},
		config,
	)
	return server.Serve(tlsListener)
}

// GenerateCert self signed with CA for use with TLS
func GenerateCert(organization string, hosts []string) ([]byte, []byte, error) {
	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf(`Error while generating cert key: %v`, err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, fmt.Errorf(`Error while generating serial number: %v`, err)
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
		return nil, nil, fmt.Errorf(`Error while getting locals ips: %v`, err)
	}

	for _, ip := range ips {
		template.IPAddresses = append(template.IPAddresses, ip)
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &ecdsaKey.PublicKey, ecdsaKey)
	if err != nil {
		return nil, nil, fmt.Errorf(`Error while creating certificate: %v`, err)
	}

	key, err := x509.MarshalECPrivateKey(ecdsaKey)
	if err != nil {
		return nil, nil, fmt.Errorf(`Error while marshalling private key: %v`, err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: `CERTIFICATE`, Bytes: der}), pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: key}), nil
}
