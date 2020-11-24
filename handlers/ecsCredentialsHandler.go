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
	"time"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/netflix/weep/util"

	"github.com/gorilla/mux"
	"github.com/netflix/weep/creds"
	"github.com/netflix/weep/metadata"
	log "github.com/sirupsen/logrus"
)

var credentialMap = make(map[string]*creds.AwsCredentials)

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

// getCacheSlug returns a string unique to a particular combination of a role and chain of roles to assume.
func getCacheSlug(role string, assume []string) string {
	elements := append([]string{role}, assume...)
	return strings.Join(elements, "/")
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
		util.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	vars := mux.Vars(r)
	requestedRole := vars["role"]
	cacheSlug := getCacheSlug(requestedRole, assume)
	var credentials *creds.AwsCredentials

	val, ok := credentialMap[cacheSlug]
	if ok {
		credentials = val

		// Refresh credentialResponse on demand if expired or within 10 minutes of expiry
		currentTime := time.Now()
		tm := time.Unix(credentials.Expiration, 0)
		timeToRenew := tm.Add(-10 * time.Minute)
		if currentTime.After(timeToRenew) {
			credentials, err = creds.GetCredentialsC(client, requestedRole, false, assume)
			if err != nil {
				log.Error(err)
				return
			}
		}
	} else {
		credentials, err = creds.GetCredentialsC(client, requestedRole, false, assume)
		if err != nil {
			log.Error(err)
			return
		}
		credentialMap[cacheSlug] = credentials
	}

	tm := time.Unix(credentials.Expiration, 0)

	credentialResponse := metadata.ECSMetaDataCredentialResponse{
		AccessKeyId:     fmt.Sprintf("%s", credentials.AccessKeyId),
		Expiration:      tm.UTC().Format("2006-01-02T15:04:05Z"),
		RoleArn:         credentials.RoleArn,
		SecretAccessKey: fmt.Sprintf("%s", credentials.SecretAccessKey),
		Token:           fmt.Sprintf("%s", credentials.SessionToken),
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
