package cmd

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/netflix/weep/consoleme"
	"github.com/netflix/weep/handlers"
	"github.com/netflix/weep/metadata"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	metadataRole       string
	metadataListenAddr string
	metadataListenPort int
)

func init() {
	metadataCmd.PersistentFlags().StringVar(&metadataRole, "role", "", "name of role")
	metadataCmd.PersistentFlags().StringVar(&metadataListenAddr, "listen_ip", "127.0.0.1", "IP address for metadata service to listen on")
	metadataCmd.PersistentFlags().IntVar(&metadataListenPort, "port", 9090, "port for metadata service to listen on")
	rootCmd.AddCommand(metadataCmd)
}

var metadataCmd = &cobra.Command{
	Use:   "metadata",
	Short: "Run a local Instance Metadata Service (IMDS) endpoint that serves credentials",
	RunE:  runMetadata,
}

func runMetadata(cmd *cobra.Command, args []string) error {
	if metadataRole == "" {

	}
	metadata.Role = metadataRole
	client, err := consoleme.GetClient()
	if err != nil {
		return err
	}
	ipaddress := net.ParseIP(metadataListenAddr)

	if ipaddress == nil {
		fmt.Println("Invalid IP: ", metadataListenAddr)
		os.Exit(1)
	}

	listener_addr := fmt.Sprintf("%s:%d", ipaddress, metadataListenPort)

	router := mux.NewRouter()
	router.HandleFunc("/{version}/", handlers.MetaDataServiceMiddleware(handlers.BaseVersionHandler))
	router.HandleFunc("/{version}/api/token", handlers.MetaDataServiceMiddleware(handlers.TokenHandler)).Methods("PUT")
	router.HandleFunc("/{version}/meta-data", handlers.MetaDataServiceMiddleware(handlers.BaseHandler))
	router.HandleFunc("/{version}/meta-data/", handlers.MetaDataServiceMiddleware(handlers.BaseHandler))
	router.HandleFunc("/{version}/meta-data/iam/info", handlers.MetaDataServiceMiddleware(handlers.IamInfoHandler))
	router.HandleFunc("/{version}/meta-data/iam/security-credentials/", handlers.MetaDataServiceMiddleware(handlers.RoleHandler))
	router.HandleFunc("/{version}/meta-data/iam/security-credentials/{role}", handlers.MetaDataServiceMiddleware(handlers.CredentialsHandler))
	router.HandleFunc("/{version}/dynamic/instance-identity/document", handlers.MetaDataServiceMiddleware(handlers.InstanceIdentityDocumentHandler))
	router.HandleFunc("/{path:.*}", handlers.MetaDataServiceMiddleware(handlers.CustomHandler))

	go metadata.StartMetaDataRefresh(client)

	go func() {
		log.Info("Starting weep meta-data service...")
		log.Info("Server started on: ", listener_addr)
		log.Info(http.ListenAndServe(listener_addr, router))
	}()

	// Check for interrupt signal and exit cleanly
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Print("Shutdown signal received, exiting weep meta-data service...")

	return nil
}
