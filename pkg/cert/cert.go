package cert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/pkg/tools"
)

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`cert`:  flag.String(tools.ToCamel(fmt.Sprintf(`%s%s`, prefix, `Cert`)), ``, `[tls] PEM Certificate file`),
		`key`:   flag.String(tools.ToCamel(fmt.Sprintf(`%s%s`, prefix, `Key`)), ``, `[tls] PEM Key file`),
		`hosts`: flag.String(tools.ToCamel(fmt.Sprintf(`%s%s`, prefix, `Hosts`)), `localhost`, `[tls] Self-signed certificate hosts, comma separated`),
	}
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
// from https://golang.org/src/net/http/server.go
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return tc, err
	}

	if err = tc.SetKeepAlive(true); err != nil {
		log.Printf(`Error while setting keep alive: %v`, err)
	}
	if err := tc.SetKeepAlivePeriod(3 * time.Minute); err != nil {
		log.Printf(`Error while setting keep alive period: %v`, err)
	}

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

// ListenAndServeTLS with provided certFile flag or self-signed generated certificate
// Largely inspired by https://golang.org/src/net/http/server.go
func ListenAndServeTLS(config map[string]*string, server *http.Server) error {
	if *config[`cert`] != `` {
		return server.ListenAndServeTLS(*config[`cert`], *config[`key`])
	}

	certPEMBlock, keyPEMBlock, err := GenerateCert(`ViBiOh`, strings.Split(*config[`hosts`], `,`))
	if err != nil {
		return fmt.Errorf(`Error while generating certificate: %v`, err)
	}
	log.Print(`Self-signed certificate generated`)

	addr := server.Addr
	if addr == `` {
		addr = `:https`
	}

	tlsConfig := &tls.Config{}
	tlsConfig.NextProtos = append(tlsConfig.NextProtos, `h2`)
	if !strSliceContains(tlsConfig.NextProtos, `http/1.1`) {
		tlsConfig.NextProtos = append(tlsConfig.NextProtos, `http/1.1`)
	}

	tlsConfig.Certificates = make([]tls.Certificate, 1)
	certificate, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return fmt.Errorf(`Error while getting x509 KeyPair: %v`, err)
	}
	tlsConfig.Certificates[0] = certificate

	listener, err := net.Listen(`tcp`, addr)
	if err != nil {
		return fmt.Errorf(`Error while listening: %v`, err)
	}

	tlsListener := tls.NewListener(
		tcpKeepAliveListener{listener.(*net.TCPListener)},
		tlsConfig,
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
	endDate := startDate.Add(time.Hour * 24 * 365)

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

	ips, err := tools.GetLocalIPS()
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

	return pem.EncodeToMemory(&pem.Block{Type: `CERTIFICATE`, Bytes: der}), pem.EncodeToMemory(&pem.Block{Type: `EC PRIVATE KEY`, Bytes: key}), nil
}
