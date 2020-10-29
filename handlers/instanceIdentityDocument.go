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

	log "github.com/sirupsen/logrus"

	"github.com/netflix/weep/metadata"
	"github.com/netflix/weep/util"
)

var (
	accountID string
)

func InstanceIdentityDocumentHandler(w http.ResponseWriter, r *http.Request) {

	awsArn, err := util.ArnParse(metadata.Role)

	if err != nil {
		accountID = "123456789012"
	} else {
		accountID = awsArn.AccountId
	}

	identityDocument := metadata.MetaDataInstanceIdentityDocumentResponse{
		DevpayProductCodes:      []string{},
		MarkerplaceProductCodes: []string{},
		PrivateIP:               "100.1.2.3",
		Version:                 "2017-09-30",
		InstanceID:              "i-12345",
		BillingProductCodes:     []string{},
		InstanceType:            "m5.large",
		AvailabilityZone:        "us-east-1a",
		KernelID:                "aki-fc8f11cc",
		RamdiskID:               "",
		AccountID:               accountID,
		Architecture:            "x86_64",
		ImageID:                 "ami-12345",
		PendingTime:             metadata.LastRenewal.UTC(), //.Format("2006-01-02T15:04:05Z"),
		Region:                  metadata.MetadataRegion,
	}

	b, err := json.Marshal(identityDocument)
	if err != nil {
		log.Error(err)
	}

	var out bytes.Buffer
	json.Indent(&out, b, "", "  ")
	fmt.Fprintln(w, out.String())
}
