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
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"

	"github.com/netflix/weep/errors"

	"github.com/aws/aws-sdk-go/aws/credentials"
)

// NewRefreshableProvider creates an AWS credential provider that will automatically refresh credentials
// when they are close to expiring
func NewRefreshableProvider(client HTTPClient, role, region string, assumeChain []string, noIpRestrict bool) (*RefreshableProvider, error) {
	splitRole := strings.Split(role, "/")
	roleName := splitRole[len(splitRole)-1]
	rp := &RefreshableProvider{
		RoleName:     roleName,
		RoleArn:      role,
		Region:       region,
		NoIpRestrict: noIpRestrict,
		AssumeChain:  assumeChain,
		client:       client,
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
			_, err := rp.checkAndRefresh(10)
			if err != nil {
				log.Error(err.Error())
			}
		}
	}
}

func (rp *RefreshableProvider) checkAndRefresh(threshold int) (bool, error) {
	log.Debugf("checking credentials for %s", rp.RoleName)
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
	log.Debugf("refreshing credentials for %s", rp.RoleArn)
	var err error
	var newCreds *AwsCredentials

	rp.Lock()
	defer rp.Unlock()

	log.WithFields(logrus.Fields{
		"roleName":     rp.RoleName,
		"roleArn":      rp.RoleArn,
		"noIpRestrict": rp.NoIpRestrict,
		"assumeChain":  rp.AssumeChain,
	}).Debug("requesting new credentials")
	newCreds, err = GetCredentialsC(rp.client, rp.RoleArn, rp.NoIpRestrict, rp.AssumeChain)
	if err != nil {
		if err == errors.MutualTLSCertNeedsRefreshError {
			log.Error(err)
			// The http.Client, with the best of intentions, will hold the connection open,
			// meaning that an auto-updated cert won't be used by the client.
			rp.client.CloseIdleConnections()
			return fmt.Errorf(viper.GetString("mtls_settings.old_cert_message"))
		} else {
			return err
		}
	}

	rp.Expiration = newCreds.Expiration
	rp.value.AccessKeyID = newCreds.AccessKeyId
	rp.value.SessionToken = newCreds.SessionToken
	rp.value.SecretAccessKey = newCreds.SecretAccessKey
	rp.value.AccessKeyID = newCreds.AccessKeyId
	rp.LastRefreshed = Time(time.Now())
	if newCreds.RoleArn != "" {
		// We favor the role ARN from ConsoleMe over the one from the user, which could just be a search string.
		rp.RoleArn = newCreds.RoleArn
	}
	if rp.value.ProviderName == "" {
		rp.value.ProviderName = "WeepRefreshableProvider"
	}
	log.Debugf("successfully refreshed credentials for %s", rp.RoleArn)
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
