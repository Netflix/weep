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

package util

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/netflix/weep/errors"
	"github.com/netflix/weep/logging"
)

var log = logging.GetLogger()

type AwsArn struct {
	Arn               string
	Partition         string
	Service           string
	Region            string
	AccountId         string
	ResourceType      string
	Resource          string
	ResourceDelimiter string
}

type ErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

func validate(arn string, pieces []string) error {
	if len(pieces) < 6 {
		return errors.InvalidArn
	}
	return nil
}

func ArnParse(arn string) (*AwsArn, error) {
	pieces := strings.SplitN(arn, ":", 6)

	if err := validate(arn, pieces); err != nil {
		return nil, err
	}

	components := &AwsArn{
		Arn:       pieces[0],
		Partition: pieces[1],
		Service:   pieces[2],
		Region:    pieces[3],
		AccountId: pieces[4],
	}
	if n := strings.Count(pieces[5], ":"); n > 0 {
		components.ResourceDelimiter = ":"
		resourceParts := strings.SplitN(pieces[5], ":", 2)
		components.ResourceType = resourceParts[0]
		components.Resource = resourceParts[1]
	} else {
		if m := strings.Count(pieces[5], "/"); m == 0 {
			components.Resource = pieces[5]
		} else {
			components.ResourceDelimiter = "/"
			resourceParts := strings.SplitN(pieces[5], "/", 2)
			components.ResourceType = resourceParts[0]
			components.Resource = resourceParts[1]
		}
	}
	return components, nil
}

func (a AwsArn) ArnString() string {
	return fmt.Sprintf("%s:%s:%s:%s:%s:%s/%s", a.Arn, a.Partition, a.Service, a.Region, a.AccountId, a.ResourceType, a.Resource)
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// WriteError writes a status code and plaintext error to the provided http.ResponseWriter.
// The error is written as plaintext so AWS SDKs will display it inline with an error message.
func WriteError(w http.ResponseWriter, message string, status int) {
	log.Debugf("writing HTTP error response: %s", message)
	w.WriteHeader(status)
	_, err := w.Write([]byte(message))
	if err != nil {
		log.Errorf("could not write error response: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Attempt to open a link in browser, if supported
func OpenLink(link string) error {
	var openUrlCommand []string = nil
	switch runtime.GOOS {
	case "darwin":
		openUrlCommand = []string{"open"}
	case "linux":
		if isWSL() {
			openUrlCommand = []string{"cmd.exe", "/C", "start"}
		} else {
			openUrlCommand = []string{"xdg-open"}
		}
	case "windows":
		// This is unsupported until we find a safer way to run the open command in Windows.
		return errors.BrowserOpenError
	}

	if openUrlCommand != nil {
		cmd := exec.Command(openUrlCommand[0], append(openUrlCommand[1:], link)...)
		err := cmd.Start()
		if err == nil {
			err = cmd.Wait()
		}
		if err != nil {
			return err
		} else {
			log.Infoln("Link opened in a new browser window.")
		}
	} else {
		return errors.BrowserOpenError
	}
	return nil
}

func isWSL() bool {
	if FileExists("/proc/sys/kernel/osrelease") {
		if osrelease, err := ioutil.ReadFile("/proc/sys/kernel/osrelease"); err == nil {
			if strings.Contains(strings.ToLower(string(osrelease)), "microsoft") {
				return true
			}
		}
	}
	return false
}
