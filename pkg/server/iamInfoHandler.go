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
	"net/http"

	"github.com/netflix/weep/pkg/logging"

	"github.com/netflix/weep/pkg/cache"
	"github.com/netflix/weep/pkg/util"
)

func IamInfoHandler(w http.ResponseWriter, r *http.Request) {
	rawArn := cache.GlobalCache.DefaultArn()
	awsArn, _ := util.ArnParse(rawArn)

	awsArn.ResourceType = "instance-profile"

	iamInfo := MetaDataIamInfoResponse{
		Code:               "Success",
		LastUpdated:        cache.GlobalCache.DefaultLastUpdated(),
		InstanceProfileARN: awsArn.ArnString(),
		InstanceProfileID:  "AIPAI",
	}

	err := json.NewEncoder(w).Encode(iamInfo)
	if err != nil {
		logging.Log.Errorf("failed to write response: %v", err)
	}
}
