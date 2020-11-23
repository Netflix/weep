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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/netflix/weep/creds"
	"github.com/netflix/weep/metadata"
	log "github.com/sirupsen/logrus"
)

var credentialMap = make(map[string]creds.AwsCredentials)

func ECSMetadataServiceCredentialsHandler(w http.ResponseWriter, r *http.Request) {
	var client, err = creds.GetClient()
	if err != nil {
		log.Error(err)
		return
	}
	vars := mux.Vars(r)
	requestedRole := vars["role"]
	var Credentials creds.AwsCredentials

	val, ok := credentialMap[requestedRole]
	if ok {
		Credentials = val

		// Refresh credentials on demand if expired or within 10 minutes of expiry
		currentTime := time.Now()
		tm := time.Unix(Credentials.Expiration, 0)
		timeToRenew := tm.Add(-10 * time.Minute)
		if currentTime.After(timeToRenew) {
			Credentials, err = client.GetRoleCredentials(requestedRole, false)
			if err != nil {
				log.Error(err)
				return
			}
		}
	} else {
		Credentials, err = client.GetRoleCredentials(requestedRole, false)
		if err != nil {
			log.Error(err)
			return
		}
		credentialMap[requestedRole] = Credentials
	}

	tm := time.Unix(Credentials.Expiration, 0)

	credentials := metadata.ECSMetaDataCredentialResponse{
		AccessKeyId:     fmt.Sprintf("%s", Credentials.AccessKeyId),
		Expiration:      tm.UTC().Format("2006-01-02T15:04:05Z"),
		RoleArn:         Credentials.RoleArn,
		SecretAccessKey: fmt.Sprintf("%s", Credentials.SecretAccessKey),
		Token:           fmt.Sprintf("%s", Credentials.SessionToken),
	}

	b, err := json.Marshal(credentials)
	if err != nil {
		log.Error(err)
	}
	var out bytes.Buffer
	json.Indent(&out, b, "", "  ")
	fmt.Fprintln(w, out.String())
}
