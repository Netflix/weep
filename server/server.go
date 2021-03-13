package server

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/netflix/weep/cache"
	"github.com/netflix/weep/creds"
)

func Run(host string, port int, role, region string, shutdown chan os.Signal) error {
	ipaddress := net.ParseIP(host)

	if ipaddress == nil {
		return fmt.Errorf("invalid IP: %s", host)
	}

	listenAddr := fmt.Sprintf("%s:%d", ipaddress, port)

	router := mux.NewRouter()
	router.HandleFunc("/healthcheck", HealthcheckHandler)

	if role != "" {
		log.Infof("Configuring weep IMDS service for role %s", role)
		client, err := creds.GetClient(region)
		if err != nil {
			return err
		}
		err = cache.GlobalCache.SetDefault(client, role, region, make([]string, 0))
		if err != nil {
			return err
		}
		router.HandleFunc("/{version}/", CredentialServiceMiddleware(BaseVersionHandler))
		router.HandleFunc("/{version}/api/token", CredentialServiceMiddleware(TokenHandler)).Methods("PUT")
		router.HandleFunc("/{version}/meta-data", CredentialServiceMiddleware(BaseHandler))
		router.HandleFunc("/{version}/meta-data/", CredentialServiceMiddleware(BaseHandler))
		router.HandleFunc("/{version}/meta-data/iam/info", CredentialServiceMiddleware(IamInfoHandler))
		router.HandleFunc("/{version}/meta-data/iam/security-credentials/", CredentialServiceMiddleware(RoleHandler))
		router.HandleFunc("/{version}/meta-data/iam/security-credentials/{role}", CredentialServiceMiddleware(IMDSHandler))
		router.HandleFunc("/{version}/dynamic/instance-identity/document", CredentialServiceMiddleware(InstanceIdentityDocumentHandler))
	}

	router.HandleFunc("/ecs/{role:.*}", CredentialServiceMiddleware(getCredentialHandler(region)))
	router.HandleFunc("/{path:.*}", CredentialServiceMiddleware(CustomHandler))

	go func() {
		log.Info("Starting weep on ", listenAddr)
		log.Info(http.ListenAndServe(listenAddr, router))
	}()

	// Check for interrupt signal and exit cleanly
	<-shutdown
	log.Print("Shutdown signal received, stopping server...")
	return nil
}
