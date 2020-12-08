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
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
)

type AwsCredentials struct {
	AccessKeyId     string    `json:"AccessKeyId"`
	SecretAccessKey string    `json:"SecretAccessKey"`
	SessionToken    string    `json:"SessionToken"`
	Expiration      time.Time `json:"Expiration"`
	RoleArn         string    `json:"RoleArn"`
}

type RefreshableProvider struct {
	value         credentials.Value
	mu            sync.RWMutex
	client        *Client
	retries       int
	retryDelay    int
	Expiration    time.Time
	LastRefreshed time.Time
	Region        string
	Role          string
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
	Credentials *AwsCredentials `json:"Credentials"`
}

type ConsolemeCredentialRequestType struct {
	RequestedRole   string `json:"requested_role"`
	NoIpRestriciton bool   `json:"no_ip_restrictions"`
}

type ConsolemeCredentialErrorMessageType struct {
	Code          string `json:"code"`
	Message       string `json:"message"`
	RequestedRole string `json:"requested_role"`
	Exception     string `json:"exception"`
	RequestID     string `json:"request_id"`
}
