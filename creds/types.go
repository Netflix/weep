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
	"strconv"
	"sync"
	"time"

	"github.com/netflix/weep/metadata"

	"github.com/aws/aws-sdk-go/aws/credentials"
)

type AwsCredentials struct {
	AccessKeyId     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
	Expiration      Time   `json:"Expiration"`
	RoleArn         string `json:"RoleArn"`
}

type RefreshableProvider struct {
	sync.RWMutex
	value         credentials.Value
	client        HTTPClient
	retries       int
	retryDelay    int
	Expiration    Time
	LastRefreshed Time
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
	Credentials *AwsCredentials `json:"Credentials"`
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

type Time time.Time

// MarshalJSON is used to convert the timestamp to JSON
func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(Time(t).Unix(), 10)), nil
}

// UnmarshalJSON is used to convert the timestamp from JSON
func (t *Time) UnmarshalJSON(s []byte) (err error) {
	r := string(s)
	q, err := strconv.ParseInt(r, 10, 64)
	if err != nil {
		return err
	}
	*(*time.Time)(t) = time.Unix(q, 0)
	return nil
}

// Add returns t with the provided duration added to it.
func (t Time) Add(d time.Duration) time.Time {
	return time.Time(t).Add(d)
}

// Unix returns t as a Unix time, the number of seconds elapsed
// since January 1, 1970 UTC. The result does not depend on the
// location associated with t.
func (t Time) Unix() int64 {
	return time.Time(t).Unix()
}

func (t Time) UTC() time.Time {
	return time.Time(t).UTC()
}

// Format returns t as a timestamp string with the provided layout.
func (t Time) Format(layout string) string {
	return time.Time(t).Format(layout)
}

// Time returns the JSON time as a time.Time instance in UTC
func (t Time) Time() time.Time {
	return time.Time(t).UTC()
}

// String returns t as a formatted string
func (t Time) String() string {
	return t.Time().String()
}

type Credentials struct {
	Role                string
	NoIpRestrict        bool
	metaDataCredentials *AwsCredentials
	MetadataRegion      string
	LastRenewal         Time
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
