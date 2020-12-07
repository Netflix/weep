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
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	log "github.com/sirupsen/logrus"
)

func NewRefreshableProvider(client *Client, role, region string, assumeChain []string, noIpRestrict bool) (*RefreshableProvider, error) {
	rp := &RefreshableProvider{
		Role:         role,
		Region:       region,
		NoIpRestrict: noIpRestrict,
		AssumeChain:  assumeChain,
		client:       client,
	}
	rp.refresh()
	// kick off a goroutine to automatically refresh creds
	go AutoRefresh(rp)
	return rp, nil
}

func AutoRefresh(rp *RefreshableProvider) {
	// we'll check the creds every minute to see if they're close to expiring
	ticker := time.NewTicker(time.Minute)

	for {
		select {
		case _ = <-ticker.C:
			log.Debugf("checking credentials for %s", rp.Role)
			// refresh creds if we're within 10 minutes of them expiring
			thresh := rp.Expiration.Add(-10 * time.Minute)
			if time.Now().After(thresh) {
				rp.refresh()
			}
		}
	}
}

func (rp *RefreshableProvider) refresh() {
	log.Debugf("refreshing credentials for %s", rp.Role)
	var err error
	var newCreds *AwsCredentials
	retryDelay := 5 * time.Second

	rp.mu.Lock()
	defer rp.mu.Unlock()

	for i := 0; i < 5; i++ {
		newCreds, err = GetCredentialsC(rp.client, rp.Role, rp.NoIpRestrict, rp.AssumeChain)
		if err != nil {
			log.Errorf("failed to get refreshed credentials: %s", err.Error())
			time.Sleep(retryDelay)
		} else {
			break
		}
	}
	if newCreds == nil {
		log.Fatal("Unable to retrieve credentials from ConsoleMe")
		os.Exit(1)
	}

	rp.Expiration = time.Unix(newCreds.Expiration, 0)
	rp.value.AccessKeyID = newCreds.AccessKeyId
	rp.value.SessionToken = newCreds.SessionToken
	rp.value.SecretAccessKey = newCreds.SecretAccessKey
	rp.value.AccessKeyID = newCreds.AccessKeyId
	rp.LastRefreshed = time.Now()
	rp.RoleArn = newCreds.RoleArn
	if rp.value.ProviderName == "" {
		rp.value.ProviderName = "WeepRefreshableProvider"
	}
	log.Debugf("successfully refreshed credentials for %s", rp.Role)
}

func (rp *RefreshableProvider) Retrieve() (credentials.Value, error) {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	return rp.value, nil
}

func (rp *RefreshableProvider) IsExpired() bool {
	// we always return false because we should never have expired creds
	return false
}
