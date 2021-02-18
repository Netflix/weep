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

	"github.com/netflix/weep/metadata"
	"github.com/netflix/weep/util"
)

func IamInfoHandler(w http.ResponseWriter, r *http.Request) {

	awsArn, _ := util.ArnParse(metadata.Role)

	awsArn.ResourceType = "instance-profile"

	iamInfo := metadata.MetaDataIamInfoResponse{
		Code:               "Success",
		LastUpdated:        metadata.LastRenewal.UTC().Format("2006-01-02T15:04:05Z"),
		InstanceProfileARN: awsArn.ArnString(),
		InstanceProfileID:  "AIPAI",
	}

	b, err := json.Marshal(iamInfo)
	if err != nil {
		log.Error(err)
	}
	var out bytes.Buffer
	json.Indent(&out, b, "", "  ")
	fmt.Fprintln(w, out.String())
}
