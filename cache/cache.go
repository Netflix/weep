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

package cache

import (
	"strings"
	"sync"

	"github.com/netflix/weep/logging"

	"github.com/netflix/weep/creds"
	"github.com/netflix/weep/errors"
	"github.com/sirupsen/logrus"
)

var GlobalCache CredentialCache
var log = logging.GetLogger()

type CredentialCache struct {
	sync.RWMutex
	RoleCredentials map[string]*creds.RefreshableProvider
	DefaultRole     string
}

func init() {
	GlobalCache = CredentialCache{
		RoleCredentials: make(map[string]*creds.RefreshableProvider),
	}
}

// getCacheSlug returns a string unique to a particular combination of a role and chain of roles to assume.
func getCacheSlug(role string, assume []string) string {
	elements := append([]string{role}, assume...)
	return strings.Join(elements, "/")
}

func (cc *CredentialCache) Get(searchString string, assumeChain []string) (*creds.RefreshableProvider, error) {
	log.WithFields(logrus.Fields{
		"searchString": searchString,
		"assumeChain":  assumeChain,
	}).Info("retrieving credentials")
	c, ok := cc.get(getCacheSlug(searchString, assumeChain))
	if ok {
		log.Debugf("found credentials for %s in cache", searchString)
		return c, nil
	}
	return nil, errors.NoCredentialsFoundInCache
}

func (cc *CredentialCache) GetOrSet(client creds.HTTPClient, role, region string, assumeChain []string) (*creds.RefreshableProvider, error) {
	c, err := cc.Get(role, assumeChain)
	if err == nil {
		return c, nil
	}
	log.Debugf("no credentials for %s in cache, creating", role)

	c, err = cc.set(client, role, region, assumeChain)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (cc *CredentialCache) SetDefault(client creds.HTTPClient, role, region string, assumeChain []string) error {
	_, err := cc.set(client, role, region, assumeChain)
	if err != nil {
		return err
	}
	cc.DefaultRole = getCacheSlug(role, assumeChain)
	return nil
}

func (cc *CredentialCache) GetDefault() (*creds.RefreshableProvider, error) {
	if cc.DefaultRole == "" {
		return nil, errors.NoDefaultRoleSet
	}
	c, ok := cc.get(cc.DefaultRole)
	if ok {
		return c, nil
	}
	return nil, errors.NoCredentialsFoundInCache
}

func (cc *CredentialCache) DefaultLastUpdated() string {
	c, err := cc.GetDefault()
	if err != nil {
		log.Debugf("cannot get last updated time of default creds: %v", err)
		return ""
	}
	return c.LastRefreshed.UTC().Format("2006-01-02T15:04:05Z")
}

func (cc *CredentialCache) DefaultArn() string {
	c, err := cc.GetDefault()
	if err != nil {
		log.Debugf("cannot get arn of default creds: %v", err)
		return ""
	}
	return c.RoleArn
}

func (cc *CredentialCache) get(slug string) (*creds.RefreshableProvider, bool) {
	cc.RLock()
	defer cc.RUnlock()
	c, ok := cc.RoleCredentials[slug]
	return c, ok
}

func (cc *CredentialCache) set(client creds.HTTPClient, role, region string, assumeChain []string) (*creds.RefreshableProvider, error) {
	c, err := creds.NewRefreshableProvider(client, role, region, assumeChain, false)
	if err != nil {
		return nil, err
	}
	cc.Lock()
	defer cc.Unlock()
	cc.RoleCredentials[getCacheSlug(role, assumeChain)] = c
	return c, nil
}
