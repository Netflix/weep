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
	"time"

	"github.com/spf13/viper"

	"github.com/netflix/weep/errors"

	"github.com/aws/aws-sdk-go/aws/credentials"
)

// NewRefreshableProvider creates an AWS credential provider that will automatically refresh credentials
// when they are close to expiring
func NewRefreshableProvider(client HTTPClient, role, region string, assumeChain []string, noIpRestrict bool) (*RefreshableProvider, error) {
	rp := &RefreshableProvider{
		Role:         role,
		Region:       region,
		NoIpRestrict: noIpRestrict,
		AssumeChain:  assumeChain,
		client:       client,
		retries:      5,
		retryDelay:   5,
	}
	err := rp.refresh()
	if err != nil {
		return nil, err
	}
	// kick off a goroutine to automatically refresh creds
	go rp.AutoRefresh()
	return rp, nil
}

func (rp *RefreshableProvider) AutoRefresh() {
	// we'll check the creds every minute to see if they're close to expiring
	ticker := time.NewTicker(time.Minute)

	for {
		select {
		case _ = <-ticker.C:
			log.Debugf("checking credentials for %s", rp.Role)
			_, err := rp.checkAndRefresh(10)
			if err != nil {
				log.Error(err.Error())
			}
		}
	}
}

func (rp *RefreshableProvider) checkAndRefresh(threshold int) (bool, error) {
	log.Debugf("checking credentials for %s", rp.Role)
	// refresh creds if we're within 10 minutes of them expiring
	diff := time.Duration(threshold*-1) * time.Minute
	thresh := rp.Expiration.Add(diff)
	if time.Now().After(thresh) {
		err := rp.refresh()
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func (rp *RefreshableProvider) refresh() error {
	log.Debugf("refreshing credentials for %s", rp.Role)
	var err error
	var newCreds *AwsCredentials
	retryDelay := time.Duration(rp.retryDelay) * time.Second

	rp.Lock()
	defer rp.Unlock()

RetryLoop:
	for i := 0; i < rp.retries; i++ {
		newCreds, err = GetCredentialsC(rp.client, rp.Role, rp.NoIpRestrict, rp.AssumeChain)
		switch err {
		case nil:
			// Everything is happy, so we don't need to retry
			break RetryLoop
		case errors.MutualTLSCertNeedsRefreshError:
			log.Error(viper.GetString("mtls_settings.old_cert_message"))
			// Only prep for the next request and sleep if we have remaining retries
			if i != rp.retries-1 {
				// The http.Client, with the best of intentions, will hold the connection open,
				// meaning that an auto-updated cert won't be used by the client.
				rp.client.CloseIdleConnections()
				time.Sleep(retryDelay)
			}
		case errors.MultipleMatchingRoles:
			return err
		default:
			log.Errorf("failed to get refreshed credentials: %s", err.Error())
			return err
		}
	}
	if err != nil {
		log.Errorf("Unable to retrieve credentials from ConsoleMe: %v", err)
		return err
	}

	rp.Expiration = newCreds.Expiration
	rp.value.AccessKeyID = newCreds.AccessKeyId
	rp.value.SessionToken = newCreds.SessionToken
	rp.value.SecretAccessKey = newCreds.SecretAccessKey
	rp.value.AccessKeyID = newCreds.AccessKeyId
	rp.LastRefreshed = Time(time.Now())
	rp.RoleArn = newCreds.RoleArn
	if rp.value.ProviderName == "" {
		rp.value.ProviderName = "WeepRefreshableProvider"
	}
	log.Debugf("successfully refreshed credentials for %s", rp.Role)
	return nil
}

// Retrieve returns the AWS credentials from the provider
func (rp *RefreshableProvider) Retrieve() (credentials.Value, error) {
	rp.RLock()
	defer rp.RUnlock()
	return rp.value, nil
}

// IsExpired always returns false because we should never have expired credentials
func (rp *RefreshableProvider) IsExpired() bool {
	return false
}
