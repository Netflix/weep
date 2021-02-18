package logging

import (
	"io"
	"os"
	"path/filepath"

	"github.com/kardianos/service"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = &logrus.Logger{
		Out:       os.Stderr,
		Formatter: new(logrus.TextFormatter),
		Level:     logrus.InfoLevel,
	}
}

// GetLogger returns the configured logger for use by the rest of the application.
func GetLogger() *logrus.Logger {
	return log
}

// UpdateConfig overrides the default logging settings. This function is meant to be
// used during CLI initialization to update the logger based on config file and CLI args.
func UpdateConfig(logLevel string, logFormat string, logFile string) error {
	// Set the log level and default to INFO
	switch logLevel {
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	// Set the log format.  Default to Text
	if logFormat == "json" {
		log.SetFormatter(&logrus.JSONFormatter{})
	} else {
		log.SetFormatter(&logrus.TextFormatter{})
	}

	var w io.Writer
	if logFile != "" {
		logDir := filepath.Dir(logFile)
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			// Log directory doesn't exist. Try to make it exist.
			log.Debugf("attempting to create log directory %s", logDir)
			err := os.MkdirAll(logDir, os.ModePerm)
			if err != nil {
				return errors.Wrap(err, "could not create log directory")
			}
		}
		// Since we hopefully have the directory, try to open the file
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			// No go. Bail out to stderr.
			log.SetOutput(os.Stderr)
			return errors.Wrapf(err, "could not open %s for logging, defaulting to stderr", logFile)
		} else if service.Interactive() {
			// No error opening the file, and we know that this is an interactive session.
			// Write to stderr and the file.
			w = io.MultiWriter(os.Stderr, file)
		} else {
			// No error on file, not interactive, so just write to the file.
			w = file
		}
	}
	log.SetOutput(w)
	log.Debug("logging configured")
	return nil
}
