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
)

var svcLogger service.Logger

type program struct{}

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

func RunService() {
	svcConfig := &service.Config{
		Name:        "Weep",
		DisplayName: "Weep",
		Description: "The ConsoleMe CLI",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	svcLogger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}
	err = s.Run()
	if err != nil {
		_ = svcLogger.Error(err)
	}
	done <- 0
}
