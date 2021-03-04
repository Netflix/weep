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

	"github.com/netflix/weep/cache"
	"github.com/netflix/weep/creds"

	"github.com/spf13/viper"

	"github.com/gorilla/mux"
	"github.com/netflix/weep/handlers"
	"github.com/spf13/cobra"
)

func init() {
	serveCmd.PersistentFlags().StringVarP(&metadataRegion, "region", "r", "us-east-1", "region of metadata service")
	serveCmd.PersistentFlags().StringVarP(&ecsProviderListenAddr, "listen-address", "a", "127.0.0.1", "IP address for the ECS credential provider to listen on")
	serveCmd.PersistentFlags().IntVarP(&ecsProviderListenPort, "port", "p", viper.GetInt("server.ecs_credential_provider_port"), "port for the ECS credential provider service to listen on")
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:     "serve [optional_role_name]",
	Aliases: []string{"ecs_credential_provider", "metadata", "imds"},
	Short:   serveShortHelp,
	Long:    serveLongHelp,
	RunE:    runWeepServer,
}

func runWeepServer(cmd *cobra.Command, args []string) error {
	ipaddress := net.ParseIP(ecsProviderListenAddr)

	if ipaddress == nil {
		return fmt.Errorf("invalid IP: %s", ecsProviderListenAddr)
	}

	listenAddr := fmt.Sprintf("%s:%d", ipaddress, ecsProviderListenPort)

	router := mux.NewRouter()
	router.HandleFunc("/healthcheck", handlers.HealthcheckHandler)

	var role string
	if len(args) > 0 {
		role = args[0]
	}
	if role != "" {
		log.Infof("Configuring weep IMDS service for role %s", role)
		client, err := creds.GetClient()
		if err != nil {
			return err
		}
		err = cache.GlobalCache.SetDefault(client, role, metadataRegion, make([]string, 0))
		if err != nil {
			return err
		}
		router.HandleFunc("/{version}/", handlers.CredentialServiceMiddleware(handlers.BaseVersionHandler))
		router.HandleFunc("/{version}/api/token", handlers.CredentialServiceMiddleware(handlers.TokenHandler)).Methods("PUT")
		router.HandleFunc("/{version}/meta-data", handlers.CredentialServiceMiddleware(handlers.BaseHandler))
		router.HandleFunc("/{version}/meta-data/", handlers.CredentialServiceMiddleware(handlers.BaseHandler))
		router.HandleFunc("/{version}/meta-data/iam/info", handlers.CredentialServiceMiddleware(handlers.IamInfoHandler))
		router.HandleFunc("/{version}/meta-data/iam/security-credentials/", handlers.CredentialServiceMiddleware(handlers.RoleHandler))
		router.HandleFunc("/{version}/meta-data/iam/security-credentials/{role}", handlers.CredentialServiceMiddleware(handlers.CredentialsHandler))
		router.HandleFunc("/{version}/dynamic/instance-identity/document", handlers.CredentialServiceMiddleware(handlers.InstanceIdentityDocumentHandler))
	}

	router.HandleFunc("/ecs/{role:.*}", handlers.CredentialServiceMiddleware(handlers.ECSMetadataServiceCredentialsHandler))
	router.HandleFunc("/{path:.*}", handlers.CredentialServiceMiddleware(handlers.CustomHandler))

	go func() {
		log.Info("Starting weep on ", listenAddr)
		log.Info(http.ListenAndServe(listenAddr, router))
	}()

	// Check for interrupt signal and exit cleanly
	<-shutdown
	log.Print("Shutdown signal received, stopping server...")
	return nil
}
