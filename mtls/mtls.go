package mtls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
)

// GetTLSConfig makes and returns a pointer to a tls.Config
func GetTLSConfig(certFile, keyFile, caFile string, insecure bool) (*tls.Config, error) {
	if certFile == "" || keyFile == "" || caFile == "" {
		log.Error("MTLS cert, key, or CA file not defined in configuration")
		return nil, MissingTLSConfigError
	}
	caCert, _ := ioutil.ReadFile(caFile)
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: insecure,
		RootCAs:            caCertPool,
		Certificates:       []tls.Certificate{cert},
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			// Based on the golang verification code. See https://golang.org/src/crypto/tls/handshake_client.go
			certs := make([]*x509.Certificate, len(rawCerts))
			for i, asn1Data := range rawCerts {
				cert, err := x509.ParseCertificate(asn1Data)
				if err != nil {
					return fmt.Errorf("tls: failed to parse certificate from server: %w", err)
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
	}
	return tlsConfig, nil
}

func NewHTTPClient() (*http.Client, error) {
	// Attempt to get a TLS config from the embedded configuration
	tlsConfig, err := GetEmbeddedTLSConfig()
	if err != nil && err == EmbeddedConfigDisabledError {
		log.Debug("Embedded MTLS config is disabled")
	} else if err != nil {
		return nil, err
	}

	if tlsConfig == nil {
		// We don't have an embedded TLS config, so we'll make one from the app config
		certFile := viper.GetString("mtls_settings.cert")
		keyFile := viper.GetString("mtls_settings.key")
		caFile := viper.GetString("mtls_settings.catrust")
		insecureSkipVerify := viper.GetBool("mtls_settings.insecure")
		tlsConfig, err = GetTLSConfig(certFile, keyFile, caFile, insecureSkipVerify)
		if err != nil {
			return nil, err
		}
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return client, nil
}
