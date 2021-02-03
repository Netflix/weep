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

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/netflix/weep/cache"
	"github.com/netflix/weep/util"

	"github.com/gorilla/mux"
	"github.com/netflix/weep/creds"
	"github.com/netflix/weep/metadata"
	log "github.com/sirupsen/logrus"
)

// parseAssumeRoleQuery extracts the assume query string argument, splits it on commas, validates that each element
// is an ARN, and returns a slice of ARN strings.
func parseAssumeRoleQuery(r *http.Request) ([]string, error) {
	assumeString := r.URL.Query().Get("assume")

	// Return an empty slice if we don't have an assume query string
	if assumeString == "" {
		return make([]string, 0), nil
	}

	roles := strings.Split(assumeString, ",")

	// Make sure we have valid ARNs
	for _, role := range roles {
		if !arn.IsARN(role) {
			return nil, fmt.Errorf("invalid ARN in assume query string: %s", role)
		}
	}

	return roles, nil
}

func ECSMetadataServiceCredentialsHandler(w http.ResponseWriter, r *http.Request) {
	var client, err = creds.GetClient()
	if err != nil {
		log.Error(err)
		return
	}
	assume, err := parseAssumeRoleQuery(r)
	if err != nil {
		log.Error(err)
		util.WriteError(w, err.Error(), http.StatusBadRequest)
		return
	}
	vars := mux.Vars(r)
	requestedRole := vars["role"]

	cached, err := cache.GlobalCache.GetOrSet(client, requestedRole, "", assume)
	if err != nil {
		// TODO: handle error better and return a helpful response/status
		log.Errorf("failed to get credentials: %s", err.Error())
		return
	}
	cachedCredentials, err := cached.Retrieve()
	if err != nil {
		// TODO: handle error better and return a helpful response/status
		log.Errorf("failed to get credentials: %s", err.Error())
		return
	}

	credentialResponse := metadata.ECSMetaDataCredentialResponse{
		AccessKeyId:     fmt.Sprintf("%s", cachedCredentials.AccessKeyID),
		Expiration:      cached.Expiration.UTC().Format("2006-01-02T15:04:05Z"),
		RoleArn:         cached.RoleArn,
		SecretAccessKey: fmt.Sprintf("%s", cachedCredentials.SecretAccessKey),
		Token:           fmt.Sprintf("%s", cachedCredentials.SessionToken),
	}

	b, err := json.Marshal(credentialResponse)
	if err != nil {
		log.Error(err)
	}
	_, err = w.Write(b)
	if err != nil {
		log.Errorf("failed to write HTTP response: %s", err)
	}
}
