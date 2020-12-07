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

package metadata

import (
	"sync"
	"time"

	"github.com/netflix/weep/creds"
)

type Credentials struct {
	Role                string
	NoIpRestrict        bool
	metaDataCredentials *creds.AwsCredentials
	MetadataRegion      string
	LastRenewal         time.Time
	mu                  sync.Mutex
}

type MetaDataCredentialResponse struct {
	Code            string
	LastUpdated     string
	Type            string
	AccessKeyId     string
	SecretAccessKey string
	Token           string
	Expiration      string
}

type ECSMetaDataCredentialResponse struct {
	AccessKeyId     string
	SecretAccessKey string
	Token           string
	Expiration      string
	RoleArn         string
}

type MetaDataIamInfoResponse struct {
	Code               string `json:"Code"`
	LastUpdated        string `json:"LastUpdated"`
	InstanceProfileARN string `json:"InstanceProfileArn"`
	InstanceProfileID  string `json:"InstanceProfileId"`
}

type MetaDataInstanceIdentityDocumentResponse struct {
	DevpayProductCodes      []string  `json:"devpayProductCodes"`
	MarkerplaceProductCodes []string  `json:"marketplaceProductCodes"`
	PrivateIP               string    `json:"privateIp"`
	Version                 string    `json:"version"`
	InstanceID              string    `json:"instanceId"`
	BillingProductCodes     []string  `json:"billingProducts"`
	InstanceType            string    `json:"instanceType"`
	AvailabilityZone        string    `json:"availabilityZone"`
	KernelID                string    `json:"kernelId"`
	RamdiskID               string    `json:"ramdiskId"`
	AccountID               string    `json:"accountId"`
	Architecture            string    `json:"architecture"`
	ImageID                 string    `json:"imageId"`
	PendingTime             time.Time `json:"pendingTime"`
	Region                  string    `json:"region"`
}
