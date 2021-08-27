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
	"fmt"
	"net/http"
	"strconv"

	"github.com/netflix/weep/pkg/session"
	"github.com/netflix/weep/pkg/util"

	"github.com/sirupsen/logrus"
)

func TokenHandler(w http.ResponseWriter, r *http.Request) {
	ttlString := r.Header.Get("X-aws-ec2-metadata-token-ttl-seconds")
	ttlSeconds, err := strconv.Atoi(ttlString)
	log.WithFields(logrus.Fields{
		"ttlSeconds": ttlSeconds,
	}).Debug("generating IMDSv2 token")
	if err != nil {
		util.WriteError(w, "bad request", http.StatusBadRequest)
		return
	}
	token := session.GenerateToken("", ttlSeconds)
	fmt.Fprint(w, token)
}
