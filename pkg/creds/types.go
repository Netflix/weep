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

package creds

import (
	"sync"

	"github.com/netflix/weep/pkg/aws"
	"github.com/netflix/weep/pkg/metadata"
	"github.com/netflix/weep/pkg/types"

	"github.com/aws/aws-sdk-go/aws/credentials"
)

type RefreshableProvider struct {
	sync.RWMutex
	value         credentials.Value
	client        HTTPClient
	retries       int
	retryDelay    int
	Expiration    types.Time
	LastRefreshed types.Time
	Region        string
	RoleName      string
	RoleArn       string
	NoIpRestrict  bool
	AssumeChain   []string
}

type CredentialProcess struct {
	Version         int    `json:"Version"`
	AccessKeyId     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
	Expiration      string `json:"Expiration"`
}

type ConsolemeCredentialResponseType struct {
	Credentials *aws.Credentials `json:"Credentials"`
}

type ConsolemeCredentialRequestType struct {
	RequestedRole  string                 `json:"requested_role"`
	NoIpRestricton bool                   `json:"no_ip_restrictions"`
	Metadata       *metadata.InstanceInfo `json:"metadata,omitempty"`
}

type ConsoleMeCredentialRequestMetadata struct {
}

type ConsolemeCredentialErrorMessageType struct {
	Code          string `json:"code"`
	Message       string `json:"message"`
	RequestedRole string `json:"requested_role"`
	Exception     string `json:"exception"`
	RequestID     string `json:"request_id"`
}

type Credentials struct {
	Role                string
	NoIpRestrict        bool
	metaDataCredentials *Credentials
	MetadataRegion      string
	LastRenewal         types.Time
	mu                  sync.Mutex
}

// ConsolemeWebResponse represents the response structure of ConsoleMe's model WebResponse
type ConsolemeWebResponse struct {
	Status      string            `json:"status"`
	Reason      string            `json:"reason"`
	RedirectURL string            `json:"redirect_url"`
	StatusCode  int               `json:"status_code"`
	Message     string            `json:"message"`
	Errors      []string          `json:"errors"`
	Data        map[string]string `json:"data"`
}
