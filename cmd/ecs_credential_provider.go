/*
 * Copyright 2020 Netflix, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/viper"

	"github.com/gorilla/mux"
	"github.com/netflix/weep/handlers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	ecsCredentialProvider.PersistentFlags().StringVarP(&metadataListenAddr, "listen-address", "a", "127.0.0.1", "IP address for the ECS credential provider to listen on")
	ecsCredentialProvider.PersistentFlags().IntVarP(&metadataListenPort, "port", "p", viper.GetInt("server.ecs_credential_provider_port"), "port for the ECS credential provider service to listen on")
	rootCmd.AddCommand(ecsCredentialProvider)
}

var ecsCredentialProvider = &cobra.Command{
	Use:   "ecs_credential_provider",
	Short: "Run a local ECS Credential Provider endpoint that serves and caches credentials for roles on demand",
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
	router.HandleFunc("/ecs/{role:.*}", handlers.MetaDataServiceMiddleware(handlers.ECSMetadataServiceCredentialsHandler))
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
	log.Print("Shutdown signal received, exiting weep...")

	return nil
}
