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

package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/netflix/weep/cache"
)

func RoleHandler(w http.ResponseWriter, r *http.Request) {
	defaultRole, err := cache.GlobalCache.GetDefault()
	if err != nil {
		fmt.Fprint(w, "error")
		return
	}
	fmt.Fprint(w, defaultRole.Role)
}

func IMDSHandler(w http.ResponseWriter, r *http.Request) {

	c, err := cache.GlobalCache.GetDefault()
	if err != nil {
		log.Errorf("could not get credentials from cache: %e", err)
	}
	credentials, err := c.Retrieve()
	if err != nil {
		log.Errorf("could not get credentials: %e", err)
	}

	credentialResponse := MetaDataCredentialResponse{
		Code:            "Success",
		LastUpdated:     c.LastRefreshed.UTC().Format("2006-01-02T15:04:05Z"),
		Type:            "AWS-HMAC",
		AccessKeyId:     fmt.Sprintf("%s", credentials.AccessKeyID),
		SecretAccessKey: fmt.Sprintf("%s", credentials.SecretAccessKey),
		Token:           fmt.Sprintf("%s", credentials.SessionToken),
		Expiration:      c.Expiration.UTC().Format("2006-01-02T15:04:05Z"),
	}

	err = json.NewEncoder(w).Encode(credentialResponse)
	if err != nil {
		log.Errorf("failed to write response: %v", err)
	}
}
