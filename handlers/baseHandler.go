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
	"fmt"
	"net/http"

	"github.com/netflix/weep/logging"
)

var log = logging.GetLogger()

func BaseHandler(w http.ResponseWriter, r *http.Request) {

	baseMetadata := `ami-id
ami-launch-index
ami-manifest-path
block-device-mapping/
hostname
iam/
instance-action
instance-id
instance-type
kernel-id
local-hostname
local-ipv4
mac
metrics/
network/
placement/
profile
public-keys/
reservation-id
security-groups
services/`

	fmt.Fprint(w, baseMetadata)
}

func BaseVersionHandler(w http.ResponseWriter, r *http.Request) {

	baseVersionPath := `dynamic
meta-data
user-data`

	fmt.Fprintln(w, baseVersionPath)
}
