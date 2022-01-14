package logging

import (
	"io"
	"os"
	"path/filepath"

	"github.com/kardianos/service"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var Log *logrus.Entry

var customLoggerRegistered bool

func init() {
	Log = &logrus.Entry{Logger: &logrus.Logger{
		Out:       os.Stderr,
		Formatter: new(logrus.TextFormatter),
		Level:     logrus.InfoLevel,
	},
	}
	customLoggerRegistered = false
}

// UpdateConfig overrides the default logging settings. This function is meant to be
// used during CLI initialization to update the logger based on config file and CLI args.
func UpdateConfig(logLevel string, logFormat string, logFile string) error {
	// Set the Log level and default to WARN
	switch logLevel {
	case "error":
		Log.Logger.SetLevel(logrus.ErrorLevel)
	case "debug":
		Log.Logger.SetLevel(logrus.DebugLevel)
	case "info":
		Log.Logger.SetLevel(logrus.InfoLevel)
	default:
		// only want to overwrite in the default case if it's not a custom logger
		if !customLoggerRegistered {
			Log.Logger.SetLevel(logrus.WarnLevel)
		}
	}

	// Set the Log format.  Default to Text
	if logFormat == "json" {
		Log.Logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		Log.Logger.SetFormatter(&logrus.TextFormatter{})
	}

	// Custom logger was registered, don't overwrite logger format
	if customLoggerRegistered {
		// If user has explicitly requested a log level, set output to be visible
		if logLevel != "" {
			Log.Logger.SetOutput(os.Stderr)
		}
		return nil
	}

	var w io.Writer
	if logFile != "" {
		logDir := filepath.Dir(logFile)
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			// Log directory doesn't exist. Try to make it exist.
			Log.Logger.Debugf("attempting to create Log directory %s", logDir)
			err := os.MkdirAll(logDir, os.ModePerm)
			if err != nil {
				return errors.Wrap(err, "could not create Log directory")
			}
		}
		// Since we hopefully have the directory, try to open the file
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			// No go. Bail out to stderr.
			Log.Logger.SetOutput(os.Stderr)
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
	Log.Logger.SetOutput(w)
	Log.Logger.Debug("logging configured")
	return nil
}

// RegisterLogger a custom logger
func RegisterLogger(l *logrus.Entry) {
	Log = l
	customLoggerRegistered = true
}

// LogError is a helper function that allows for errors to be logged easily
func LogError(err error, message string) {
	Log.WithError(err).Errorln(message)
}
