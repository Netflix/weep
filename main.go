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

package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/kardianos/service"
	"github.com/netflix/weep/cmd"
	log "github.com/sirupsen/logrus"
)

var svcLogger service.Logger
var done chan int

func init() {
	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)
}

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	shutdown := make(chan os.Signal, 1)
	done = make(chan int, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	cmd.Execute(shutdown, done)
}

func (p *program) Stop(s service.Service) error {
	<-done
	return nil
}

func main() {
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
}
