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
	"github.com/kardianos/service"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var svcLogger service.Logger

type program struct{}

func init() {
	weepService.Args = cobra.MinimumNArgs(1)
	rootCmd.AddCommand(weepService)
}

var weepService = &cobra.Command{
	Use:   "service [start|stop|restart|install|uninstall] [subcommand] [flags]",
	Short: "Install or control weep as a system service",
	RunE:  runWeepService,
}

//func getServiceArgsFromConfig() ([]string, error) {
//	serviceCommand := viper.GetString("service.command")
//	serviceRole := viper.GetString("service.role")
//	serviceAssume := viper.GetStringSlice("service.assume_role_chain")
//
//	return args, nil
//}

func runWeepService(cmd *cobra.Command, args []string) error {
	svcConfig := &service.Config{
		Name:        "weep",
		DisplayName: "Weep",
		Description: "The ConsoleMe CLI",
		Arguments:   args[1:],
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	errs := make(chan error, 5)
	svcLogger, err = s.Logger(errs)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			err := <-errs
			if err != nil {
				log.Error(err)
			}
		}
	}()

	if len(args[0]) > 0 {
		err := service.Control(s, args[0])
		if err != nil {
			log.Fatal(err)
		}
		done <- 0
		return nil
	}

	err = s.Run()
	if err != nil {
		_ = svcLogger.Error(err)
	}
	done <- 0

	return nil
}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	_ = rootCmd.Execute()
	done <- 0
}

func (p *program) Stop(s service.Service) error {
	<-done
	return nil
}
