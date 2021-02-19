package mtls

import (
	"crypto/tls"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/bep/debounce"
	"github.com/fsnotify/fsnotify"
)

// wrappedCertificate is a wrapper for a tls.Certificate that supports automatically
// reloading the certificate when a file change is detected.
type wrappedCertificate struct {
	sync.Mutex
	certificate *tls.Certificate
	certFile    string
	keyFile     string
}

// newWrappedCertificate initializes and returns a wrappedCertificate that will auto-
// refresh on cert/key file changes.
func newWrappedCertificate(certFile, keyFile string) (*wrappedCertificate, error) {
	log.WithFields(logrus.Fields{
		"certFile": certFile,
		"keyFile":  keyFile,
	}).Debug("creating wrapped certificate")
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	wc := wrappedCertificate{
		certificate: &cert,
		certFile:    certFile,
		keyFile:     keyFile,
	}
	go wc.autoRefresh()
	return &wc, nil
}

// getCertificate is a function to be used as the GetClientCertificate member of a tls.Config
func (wc *wrappedCertificate) getCertificate(clientHello *tls.CertificateRequestInfo) (*tls.Certificate, error) {
	wc.Lock()
	defer wc.Unlock()

	return wc.certificate, nil
}

// reloadCertificate replaces certificate with a new keypair loaded in from the filesystem.
func (wc *wrappedCertificate) reloadCertificate() {
	log.Debug("reloading mTLS certificate")
	wc.Lock()
	defer wc.Unlock()
	cert, err := tls.LoadX509KeyPair(wc.certFile, wc.keyFile)
	if err != nil {
		log.Errorf("could not reload mTLS cert: %v", err)
		return
	}
	wc.certificate = &cert
}

func (wc *wrappedCertificate) autoRefresh() {
	log.Debug("starting mTLS cert auto-refresher")

	// create the fsnotify watcher that we'll use to monitor the cert and key files
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// this channel will block the autoRefresh function from returning until
	// it's time for the program to exit (i.e. on an OS interrupt)
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	// fsnotify gives us a buuuunch of events when a refresh is done, so this
	// is here to cut down on some churn
	debounced := debounce.New(100 * time.Millisecond)

	// spin off a goroutine to handle fsnotify events
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					log.Warn("problem with mTLS file watcher")
					return
				}
				log.Infof("event received: %v", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					debounced(func() { wc.reloadCertificate() })
				}
			case watcherError, ok := <-watcher.Errors:
				if !ok {
					log.Warn("problem with mTLS file watcher")
					return
				}
				log.Error(watcherError)
			}
		}
	}()

	// add cert and key files to the watcher
	for _, file := range []string{wc.certFile, wc.keyFile} {
		err = watcher.Add(file)
		if err != nil {
			log.Fatal(err)
		}
	}

	<-interrupt
	log.Debug("stopping mTLS cert auto-refresher")
}
