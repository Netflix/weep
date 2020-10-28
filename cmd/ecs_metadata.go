package cmd

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/netflix/weep/handlers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	ecsMetadataCmd.PersistentFlags().StringVarP(&metadataListenAddr, "listen-address", "a", "127.0.0.1", "IP address for metadata service to listen on")
	ecsMetadataCmd.PersistentFlags().IntVarP(&metadataListenPort, "port", "p", 9090, "port for metadata service to listen on")
	rootCmd.AddCommand(ecsMetadataCmd)
}

var ecsMetadataCmd = &cobra.Command{
	Use:   "ecs_metadata",
	Short: "Run a local ECS Metadata Service (IMDS) endpoint that serves and caches credentials for roles on demand",
	RunE:  runEcsMetadata,
}

func runEcsMetadata(cmd *cobra.Command, args []string) error {
	ipaddress := net.ParseIP(metadataListenAddr)

	if ipaddress == nil {
		fmt.Println("Invalid IP: ", metadataListenAddr)
		os.Exit(1)
	}

	listenAddr := fmt.Sprintf("%s:%d", ipaddress, metadataListenPort)

	router := mux.NewRouter()
	router.HandleFunc("/ecs_imds/{role:.*}", handlers.MetaDataServiceMiddleware(handlers.ECSMetadataServiceCredentialsHandler))
	router.HandleFunc("/{path:.*}", handlers.MetaDataServiceMiddleware(handlers.CustomHandler))

	go func() {
		log.Info("Starting weep ECS meta-data service...")
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
