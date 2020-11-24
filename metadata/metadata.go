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
	"time"

	"github.com/netflix/weep/util"
	log "github.com/sirupsen/logrus"

	"github.com/netflix/weep/creds"
)

var (
	Role                string
	NoIpRestrict        bool
	MetaDataCredentials creds.AwsCredentials
	MetadataRegion      string
	LastRenewal         time.Time
)

func StartMetaDataRefresh(client *creds.Client) {
	retryDelay := 5 * time.Second
	retryCount := 10
	var err error
	for {
		// TODO: If 403 response,
		MetaDataCredentials, err = client.GetRoleCredentials(Role, NoIpRestrict)
		util.CheckError(err)
		Role = MetaDataCredentials.RoleArn
		if err != nil {
			log.Error(err)
			time.Sleep(retryDelay)
			if retryCount < 5 {
				continue
			} else {
				log.Fatal("Unable to retrieve credentials from ConsoleMe")
			}
		}

		expiration := time.Unix(MetaDataCredentials.Expiration, 0)

		LastRenewal = time.Now()
		timeToRenew := expiration.Add(-10 * time.Minute)
		nextRenew := timeToRenew.Sub(time.Now())
		log.Debug("meta-data: Sleeping ", nextRenew.Seconds(), " seconds until next renew")
		time.Sleep(time.Duration(nextRenew.Seconds()) * time.Second)
	}
}
