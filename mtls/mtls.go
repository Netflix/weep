package mtls

import (
	"crypto/tls"
	"crypto/x509"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
)
type CertificateParseError struct{ Cause error }

func (c CertificateParseError) Error() string {
	return "tls: failed to parse certificate from server: " + c.Cause.Error()
}

func NewHTTPClient() (*http.Client, error) {
	certFile := viper.GetString("mtls_settings.cert")
	keyFile := viper.GetString("mtls_settings.key")
	caFile := viper.GetString("mtls_settings.catrust")
	insecureSkipVerify := viper.GetBool("mtls_settings.insecure")
	if certFile == "" || keyFile == "" || caFile == "" {
		log.Fatal("MTLS cert, key, or CA file not defined in configuration")
	}
	caCert, _ := ioutil.ReadFile(caFile)
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecureSkipVerify,
				RootCAs: caCertPool,
				Certificates: []tls.Certificate{cert},
				VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
					// Based on the golang verification code. See https://golang.org/src/crypto/tls/handshake_client.go
					certs := make([]*x509.Certificate, len(rawCerts))
					for i, asn1Data := range rawCerts {
						cert, err := x509.ParseCertificate(asn1Data)
						if err != nil {
							return &CertificateParseError{err}
						}
						certs[i] = cert
					}

					opts := x509.VerifyOptions{
						Roots:         caCertPool,
						DNSName:       "",
						Intermediates: x509.NewCertPool(),
					}

					for i, cert := range certs {
						if i == 0 {
							continue
						}
						opts.Intermediates.AddCert(cert)
					}
					verifiedChains, err := certs[0].Verify(opts)
					if err != nil {
						return err
					}
					return nil
				},
			},
		},
	}

	return client, err
}
