package config

import (
	"github.com/markbates/pkger"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	EmbeddedConfigFile string // To be set by ldflags at compile time
)

// ReadEmbeddedConfig attempts to read the embedded mTLS config and create a tls.Config
func ReadEmbeddedConfig() error {
	if EmbeddedConfigFile == "" {
		return EmbeddedConfigDisabledError
	}
	f, err := pkger.Open(EmbeddedConfigFile)
	if err != nil {
		return errors.Wrap(err, "could not open embedded config")
	}
	defer f.Close()

	err = viper.ReadConfig(f)
	if err != nil {
		return errors.Wrap(err, "could not read embedded config")
	}
	return nil
}
