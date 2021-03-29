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
	"github.com/spf13/viper"

	"github.com/netflix/weep/server"
	"github.com/spf13/cobra"
)

func init() {
	serveCmd.PersistentFlags().StringVarP(&listenAddr, "listen-address", "a", viper.GetString("server.address"), "IP address for the ECS credential provider to listen on")
	serveCmd.PersistentFlags().IntVarP(&listenPort, "port", "p", viper.GetInt("server.port"), "port for the ECS credential provider service to listen on")
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
	var role string
	if len(args) > 0 {
		role = args[0]
	}
	return server.Run(listenAddr, listenPort, role, region, shutdown)
}
