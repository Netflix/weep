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
	metadataRegion     string
	metadataListenAddr string
	metadataListenPort int
)

func init() {
	metadataCmd.PersistentFlags().StringVarP(&metadataRegion, "region", "r", "us-east-1", "region of metadata service")
	metadataCmd.PersistentFlags().StringVarP(&metadataListenAddr, "listen-address", "a", "127.0.0.1", "IP address for metadata service to listen on")
	metadataCmd.PersistentFlags().IntVarP(&metadataListenPort, "port", "p", 9090, "port for metadata service to listen on")
	rootCmd.AddCommand(metadataCmd)
}

var metadataCmd = &cobra.Command{
	Use:   "metadata [role_name]",
	Short: "Run a local Instance Metadata Service (IMDS) endpoint that serves credentials",
	Args:  cobra.ExactArgs(1),
	RunE:  runMetadata,
}

func runMetadata(cmd *cobra.Command, args []string) error {
	metadataRole = args[0]
	metadata.Role = metadataRole
	metadata.MetadataRegion = metadataRegion
	client, err := consoleme.GetClient()
	if err != nil {
		return err
	}
	ipaddress := net.ParseIP(metadataListenAddr)

	if ipaddress == nil {
		fmt.Println("Invalid IP: ", metadataListenAddr)
		os.Exit(1)
	}

	listenAddr := fmt.Sprintf("%s:%d", ipaddress, metadataListenPort)

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
		log.Info("Server started on: ", listenAddr)
		log.Info(http.ListenAndServe(listenAddr, router))
	}()

	// Check for interrupt signal and exit cleanly
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Print("Shutdown signal received, exiting weep meta-data service...")

	return nil
}
