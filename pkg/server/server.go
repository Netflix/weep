package server

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/netflix/weep/pkg/logging"

	"github.com/netflix/weep/pkg/cache"
	"github.com/netflix/weep/pkg/creds"

	"github.com/gorilla/mux"
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
		logging.Log.Infof("Configuring weep IMDS service for role %s", role)
		client, err := creds.GetClient(region)
		if err != nil {
			return err
		}
		err = cache.GlobalCache.SetDefault(client, role, region, make([]string, 0))
		if err != nil {
			return err
		}

		// Unauthenticated endpoints
		router.HandleFunc("/{version}/api/token", TaskMetadataMiddleware(TokenHandler)).Methods("PUT")

		// Authenticated endpoints
		router.HandleFunc("/{version}/", InstanceMetadataMiddleware(BaseVersionHandler))
		router.HandleFunc("/{version}/meta-data", InstanceMetadataMiddleware(BaseHandler))
		router.HandleFunc("/{version}/meta-data/", InstanceMetadataMiddleware(BaseHandler))
		router.HandleFunc("/{version}/meta-data/iam/info", InstanceMetadataMiddleware(IamInfoHandler))
		// There's an extra route here to support the lack of trailing slash without the redirect that StrictSlash(true) does
		router.HandleFunc("/{version}/meta-data/iam/security-credentials", InstanceMetadataMiddleware(RoleHandler))
		router.HandleFunc("/{version}/meta-data/iam/security-credentials/", InstanceMetadataMiddleware(RoleHandler))
		router.HandleFunc("/{version}/meta-data/iam/security-credentials/{role}", InstanceMetadataMiddleware(IMDSHandler))
		router.HandleFunc("/{version}/dynamic/instance-identity/document", InstanceMetadataMiddleware(InstanceIdentityDocumentHandler))
	}

	router.HandleFunc("/ecs/{role:.*}", TaskMetadataMiddleware(getCredentialHandler(region)))
	router.HandleFunc("/{path:.*}", TaskMetadataMiddleware(NotFoundHandler))

	go func() {
		logging.Log.Info("starting weep on ", listenAddr)
		srv := &http.Server{
			ReadTimeout:       1 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       30 * time.Second,
			ReadHeaderTimeout: 2 * time.Second,
			Addr:              listenAddr,
			Handler:           router,
		}
		if err := srv.ListenAndServe(); err != nil {
			logging.Log.Fatalf("server failed: %v", err)
		}
	}()

	// Check for interrupt signal and exit cleanly
	<-shutdown
	logging.Log.Print("shutdown signal received, stopping server...")
	return nil
}
