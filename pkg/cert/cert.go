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
	"math/big"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/tools"
)

// Config of package
type Config struct {
	Cert         *string
	Key          *string
	Organization *string
	Hosts        *string
}

// Flags add flags for given prefix
func Flags(fs *flag.FlagSet, prefix string) Config {
	docPrefix := prefix
	if prefix == "" {
		docPrefix = "tls"
	}

	return Config{
		Cert:         fs.String(tools.ToCamel(fmt.Sprintf("%sCert", prefix)), "", fmt.Sprintf("[%s] PEM Certificate file", docPrefix)),
		Key:          fs.String(tools.ToCamel(fmt.Sprintf("%sKey", prefix)), "", fmt.Sprintf("[%s] PEM Key file", docPrefix)),
		Organization: fs.String(tools.ToCamel(fmt.Sprintf("%sOrganization", prefix)), "ViBiOh", fmt.Sprintf("[%s] Self-signed certificate organization", docPrefix)),
		Hosts:        fs.String(tools.ToCamel(fmt.Sprintf("%sHosts", prefix)), "localhost", fmt.Sprintf("[%s] Self-signed certificate hosts, comma separated", docPrefix)),
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
		logger.Error("%+v", errors.WithStack(err))
	}
	if err := tc.SetKeepAlivePeriod(3 * time.Minute); err != nil {
		logger.Error("%+v", errors.WithStack(err))
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
func ListenAndServeTLS(config Config, server *http.Server) error {
	cert := strings.TrimSpace(*config.Cert)
	if cert != "" {
		return server.ListenAndServeTLS(cert, strings.TrimSpace(*config.Key))
	}

	certPEMBlock, keyPEMBlock, err := GenerateFromConfig(config)
	if err != nil {
		return err
	}
	logger.Info("Self-signed certificate generated")

	addr := server.Addr
	if addr == "" {
		addr = ":https"
	}

	tlsConfig := &tls.Config{}
	tlsConfig.NextProtos = append(tlsConfig.NextProtos, "h2")
	if !strSliceContains(tlsConfig.NextProtos, "http/1.1") {
		tlsConfig.NextProtos = append(tlsConfig.NextProtos, "http/1.1")
	}

	tlsConfig.Certificates = make([]tls.Certificate, 1)
	certificate, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return errors.WithStack(err)
	}
	tlsConfig.Certificates[0] = certificate

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.WithStack(err)
	}

	tlsListener := tls.NewListener(
		tcpKeepAliveListener{listener.(*net.TCPListener)},
		tlsConfig,
	)
	return server.Serve(tlsListener)
}

// GenerateFromConfig generates certs from given config
func GenerateFromConfig(config Config) ([]byte, []byte, error) {
	return Generate(strings.TrimSpace(*config.Organization), strings.Split(strings.TrimSpace(*config.Hosts), ","))
}

// Generate self signed with CA for use with TLS
func Generate(organization string, hosts []string) ([]byte, []byte, error) {
	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, errors.WithStack(err)
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
		IsCA:                  true,
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
		return nil, nil, err
	}

	for _, ip := range ips {
		template.IPAddresses = append(template.IPAddresses, ip)
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &ecdsaKey.PublicKey, ecdsaKey)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	key, err := x509.MarshalECPrivateKey(ecdsaKey)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: key}), nil
}
