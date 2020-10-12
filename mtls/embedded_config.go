package mtls

import (
	"fmt"
	"github.com/markbates/pkger"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	EmbeddedConfigFile string
)

type Error string

func (e Error) Error() string { return string(e) }

const EmbeddedConfigDisabled = Error("embedded config is disabled")

type MTLSConfig struct {
	Enabled      bool     `yaml:"enabled"`
	InsecureSkipVerify bool `yaml:"insecure"`
	CertFilename string   `yaml:"cert_filename"`
	KeyFilename  string   `yaml:"key_filename"`
	CAFilename   string   `yaml:"ca_filename"`
	Darwin       []string `yaml:"darwin"`
	Linux        []string `yaml:"linux"`
	Windows      []string `yaml:"windows"`
}

func GetEmbeddedConfig() (string, string, string, bool, error) {
	if EmbeddedConfigFile == "" {
		return "", "", "", false, EmbeddedConfigDisabled
	}
	conf, err := ReadMTLSConfig()
	if err != nil {
		return "", "", "", false, err
	}
	dirs, err := getConfigDirs(conf)
	cert, key, ca, insecure, err := getClientCertificatePath(dirs, conf)
	if err != nil {
		return "", "", "", false, err
	}
	return cert, key, ca, insecure, nil
}

func ReadMTLSConfig() (*MTLSConfig, error) {
	var conf MTLSConfig
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

func getConfigDirs(conf *MTLSConfig) ([]string, error) {
	var mtlsDirs []string
	switch os := runtime.GOOS; os {
	case "darwin":
		mtlsDirs = conf.Darwin
	case "linux":
		mtlsDirs = conf.Linux
	case "windows":
		mtlsDirs = conf.Windows
	default:
		return nil, fmt.Errorf("running on unsupported OS %s", os)
	}
	log.Debugf("%v", mtlsDirs)

	// Replace $HOME token with home dir
	homeDir, err := getUserHome()
	if err != nil {
		return nil, fmt.Errorf("could not get user's home directory")
	}
	for i, path := range mtlsDirs {
		log.Debug(path)
		mtlsDirs[i] = strings.Replace(path, "$HOME", homeDir, -1)
	}
	log.Debugf("%v", mtlsDirs)
	return mtlsDirs, nil
}

func getClientCertificatePath(configDirs []string, conf *MTLSConfig) (string, string, string, bool, error) {
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
	return "", "", "", false, fmt.Errorf("could not find client certificates")
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

func getUserHome() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	return currentUser.HomeDir, nil
}
