package mtls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"

	"path/filepath"
	"runtime"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/netflix/weep/config"
	"github.com/netflix/weep/util"
	log "github.com/sirupsen/logrus"
)

// GetTLSConfig makes and returns a pointer to a tls.Config
func GetTLSConfig(mtlsConfig *config.MtlsSettings) (*tls.Config, error) {
	dirs, err := getTLSDirs(mtlsConfig)
	if err != nil {
		return nil, err
	}
	certFile, keyFile, caFile, insecure, err := getClientCertificatePaths(dirs, mtlsConfig)
	if err != nil {
		return nil, err
	}
	tlsConfig, err := makeTLSConfig(certFile, keyFile, caFile, insecure)
	if err != nil {
		return nil, err
	}
	return tlsConfig, nil
}

func makeTLSConfig(certFile, keyFile, caFile string, insecure bool) (*tls.Config, error) {
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
	mtlsConfig := &config.Config.MtlsSettings
	tlsConfig, err := GetTLSConfig(mtlsConfig)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return client, nil
}

// getTLSDirs returns a list of directories to search for mTLS certs based on platform
func getTLSDirs(conf *config.MtlsSettings) ([]string, error) {
	var mtlsDirs []string

	// Select config section based on platform
	switch goos := runtime.GOOS; goos {
	case "darwin":
		mtlsDirs = conf.Darwin
	case "linux":
		mtlsDirs = conf.Linux
	case "windows":
		mtlsDirs = conf.Windows
	default:
		return nil, UnsupportedOSError
	}

	// Replace $HOME token with home dir
	homeDir, err := homedir.Dir()
	if err != nil {
		return nil, HomeDirectoryError
	}
	for i, path := range mtlsDirs {
		mtlsDirs[i] = strings.Replace(path, "$HOME", homeDir, -1)
	}
	return mtlsDirs, nil
}

func getClientCertificatePaths(configDirs []string, mtlsConfig *config.MtlsSettings) (string, string, string, bool, error) {
	// If cert, key, and catrust are paths that exist, we'll just use those
	if util.FileExists(mtlsConfig.Cert) && util.FileExists(mtlsConfig.Key) && util.FileExists(mtlsConfig.CATrust) {
		return mtlsConfig.Cert, mtlsConfig.Key, mtlsConfig.CATrust, mtlsConfig.Insecure, nil
	}

	// Otherwise, get a platform-specific list of directories and look for the files there
	configDirs, err := getTLSDirs(mtlsConfig)
	if err != nil {
		return "", "", "", false, err
	}
	for _, metatronDir := range configDirs {
		certPath := filepath.Join(metatronDir, mtlsConfig.Cert)
		if !util.FileExists(certPath) {
			continue
		}

		keyPath := filepath.Join(metatronDir, mtlsConfig.Key)
		if !util.FileExists(keyPath) {
			continue
		}

		caPath := filepath.Join(metatronDir, mtlsConfig.CATrust)
		if !util.FileExists(caPath) {
			continue
		}

		return certPath, keyPath, caPath, mtlsConfig.Insecure, nil
	}
	return "", "", "", false, config.ClientCertificatesNotFoundError
}
