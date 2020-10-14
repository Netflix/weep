package mtls

import (
	"crypto/tls"
	"github.com/markbates/pkger"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	EmbeddedConfigFile string // To be set by ldflags at compile time
)

type embeddedTLSConfig struct {
	Enabled            bool     `yaml:"enabled"`
	InsecureSkipVerify bool     `yaml:"insecure"`
	CertFilename       string   `yaml:"cert_filename"`
	KeyFilename        string   `yaml:"key_filename"`
	CAFilename         string   `yaml:"ca_filename"`
	Darwin             []string `yaml:"darwin"`
	Linux              []string `yaml:"linux"`
	Windows            []string `yaml:"windows"`
}

// GetEmbeddedTLSConfig attempts to read the embedded mTLS config and create a tls.Config
func GetEmbeddedTLSConfig() (*tls.Config, error) {
	if EmbeddedConfigFile == "" {
		return nil, EmbeddedConfigDisabledError
	}
	conf, err := readEmbeddedTLSConfig()
	if err != nil {
		return nil, err
	}
	if !conf.Enabled {
		return nil, EmbeddedConfigDisabledError
	}
	dirs, err := getConfigDirs(conf)
	if err != nil {
		return nil, err
	}
	cert, key, ca, insecure, err := getClientCertificatePaths(dirs, conf)
	if err != nil {
		return nil, err
	}
	tlsConfig, err := GetTLSConfig(cert, key, ca, insecure)
	if err != nil {
		return nil, err
	}
	return tlsConfig, nil
}

func readEmbeddedTLSConfig() (*embeddedTLSConfig, error) {
	var conf embeddedTLSConfig
	f, err := pkger.Open(EmbeddedConfigFile)
	if err != nil {
		log.Errorf("could not open mtls config file: %s", EmbeddedConfigFile)
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		// TODO: handle error better
		return nil, err
	}
	fileData := make([]byte, info.Size())
	if _, err = f.Read(fileData); err != nil {
		log.Fatal("could not read mtls config, read %d bytes")
	}
	err = yaml.Unmarshal(fileData, &conf)
	if err != nil {
		// TODO: handle error better
		log.Fatal("could not load mtls config")
		return nil, err
	}
	return &conf, nil
}

// getConfigDirs returns a list of directories to search for mTLS certs based on platform
func getConfigDirs(conf *embeddedTLSConfig) ([]string, error) {
	var mtlsDirs []string

	// Select config section based on platform
	switch os := runtime.GOOS; os {
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

func getClientCertificatePaths(configDirs []string, conf *embeddedTLSConfig) (string, string, string, bool, error) {
	for _, metatronDir := range configDirs {
		certPath := filepath.Join(metatronDir, conf.CertFilename)
		if exists, err := fileExists(certPath); err != nil {
			return "", "", "", false, err
		} else if !exists {
			continue
		}

		keyPath := filepath.Join(metatronDir, conf.KeyFilename)
		if exists, err := fileExists(keyPath); err != nil {
			return "", "", "", false, err
		} else if !exists {
			continue
		}

		caPath := filepath.Join(metatronDir, conf.CAFilename)
		if exists, err := fileExists(caPath); err != nil {
			return "", "", "", false, err
		} else if !exists {
			continue
		}

		return certPath, keyPath, caPath, conf.InsecureSkipVerify, nil
	}
	return "", "", "", false, ClientCertificatesNotFoundError
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
