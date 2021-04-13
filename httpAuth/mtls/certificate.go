package mtls

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/netflix/weep/metadata"

	"github.com/sirupsen/logrus"

	"github.com/bep/debounce"
	"github.com/fsnotify/fsnotify"
)

// wrappedCertificate is a wrapper for a tls.Certificate that supports automatically
// reloading the certificate when a file change is detected.
type wrappedCertificate struct {
	sync.RWMutex
	certificate     *tls.Certificate
	x509Certificate *x509.Certificate
	certFile        string
	keyFile         string
}

// newWrappedCertificate initializes and returns a wrappedCertificate that will auto-
// refresh on cert/key file changes.
func newWrappedCertificate(certFile, keyFile string) (*wrappedCertificate, error) {
	log.WithFields(logrus.Fields{
		"certFile": certFile,
		"keyFile":  keyFile,
	}).Debug("creating wrapped certificate")

	wc := wrappedCertificate{
		certFile: certFile,
		keyFile:  keyFile,
	}
	wc.loadCertificate()
	go wc.autoRefresh()
	return &wc, nil
}

// getCertificate is a function to be used as the GetClientCertificate member of a tls.Config
func (wc *wrappedCertificate) getCertificate(clientHello *tls.CertificateRequestInfo) (*tls.Certificate, error) {
	log.Debug("getCertificate called")
	wc.RLock()
	defer wc.RUnlock()

	return wc.certificate, nil
}

func (wc *wrappedCertificate) watchExpiration() {
	expiration := wc.x509Certificate.NotAfter
	if expiration.Before(time.Now()) {
		// cert is expired, set unhealthy
	} else if expiration.Before(time.Now().Add(-6 * time.Hour)) {
		// cert is expiring soon, log warning
	}
}

// loadCertificate replaces certificate with a keypair loaded in from the filesystem.
func (wc *wrappedCertificate) loadCertificate() {
	log.Debug("reloading mTLS certificate")
	wc.Lock()
	defer wc.Unlock()
	cert, err := tls.LoadX509KeyPair(wc.certFile, wc.keyFile)
	if err != nil {
		log.Errorf("could not reload mTLS cert: %v", err)
		return
	}
	wc.certificate = &cert
	wc.x509Certificate, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		log.Errorf("could not parse x509 certificate")
	}
	wc.updateInstanceInfo()
}

func (wc *wrappedCertificate) autoRefresh() {
	log.Debug("starting mTLS cert auto-refresher")

	// create the fsnotify watcher that we'll use to monitor the cert and key files
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Errorf("mTLS cert watcher encountered an error: %v", err)
		return
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
				log.Debugf("event received: %v", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					debounced(func() { wc.loadCertificate() })
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

func (wc *wrappedCertificate) Fingerprint() string {
	fingerprintBytes := sha256.Sum256(wc.certificate.Certificate[0])
	return fmt.Sprintf("%x", fingerprintBytes)
}

func (wc *wrappedCertificate) CreateTime() time.Time {
	return wc.x509Certificate.NotBefore
}

// updateInstanceInfo makes a call to update the metadata package with the creation time
// and fingerprint of the newly-loaded certificate.
func (wc *wrappedCertificate) updateInstanceInfo() {
	metadata.SetCertInfo(wc.CreateTime(), wc.Fingerprint())
}
