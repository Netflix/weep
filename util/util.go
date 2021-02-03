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
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

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
	Error string `json:"error"`
}

func validate(arn string, pieces []string) error {
	if len(pieces) < 6 {
		return fmt.Errorf("malformed ARN: %s", arn)
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

// WriteError writes a status code and JSON-formatted error to the provided http.ResponseWriter.
func WriteError(w http.ResponseWriter, message string, status int) {
	log.Debugf("writing HTTP error response: %s", message)
	resp := ErrorResponse{Error: message}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		log.Errorf("could not marshal error response: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	_, err = w.Write(respBytes)
	if err != nil {
		log.Errorf("could not write error response: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
