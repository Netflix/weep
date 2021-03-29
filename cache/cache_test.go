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
	"testing"
	"time"

	"github.com/netflix/weep/creds"
	"github.com/netflix/weep/errors"
)

func TestCredentialCache_Get(t *testing.T) {
	cases := []struct {
		CacheContents  map[string]*creds.RefreshableProvider
		Description    string
		Role           string
		AssumeChain    []string
		ExpectedResult *creds.RefreshableProvider
		ExpectedError  error
	}{
		{
			Description:    "role not in cache",
			CacheContents:  make(map[string]*creds.RefreshableProvider),
			Role:           "a",
			AssumeChain:    []string{},
			ExpectedError:  errors.NoCredentialsFoundInCache,
			ExpectedResult: nil,
		},
		{
			Description: "role in cache",
			CacheContents: map[string]*creds.RefreshableProvider{
				"a": {RoleName: "a"},
			},
			Role:           "a",
			AssumeChain:    []string{},
			ExpectedError:  nil,
			ExpectedResult: &creds.RefreshableProvider{RoleName: "a"},
		},
		{
			Description: "role in cache with assume",
			CacheContents: map[string]*creds.RefreshableProvider{
				"a":     {RoleName: "a"},
				"a/b/c": {RoleName: "a/b/c"},
			},
			Role:           "a",
			AssumeChain:    []string{},
			ExpectedError:  nil,
			ExpectedResult: &creds.RefreshableProvider{RoleName: "a"},
		},
		{
			Description: "assume role in cache",
			CacheContents: map[string]*creds.RefreshableProvider{
				"a/b/c": {RoleName: "a/b/c"},
			},
			Role:           "a",
			AssumeChain:    []string{"b", "c"},
			ExpectedError:  nil,
			ExpectedResult: &creds.RefreshableProvider{RoleName: "a/b/c"},
		},
		{
			Description: "assume role in cache with non-assume",
			CacheContents: map[string]*creds.RefreshableProvider{
				"a":     {RoleName: "a"},
				"a/b/c": {RoleName: "a/b/c"},
			},
			Role:           "a",
			AssumeChain:    []string{"b", "c"},
			ExpectedError:  nil,
			ExpectedResult: &creds.RefreshableProvider{RoleName: "a/b/c"},
		},
	}

	for i, tc := range cases {
		t.Logf("test case %d: %s", i, tc.Description)
		testCache := CredentialCache{
			RoleCredentials: tc.CacheContents,
		}
		actualResult, actualError := testCache.Get(tc.Role, tc.AssumeChain)
		if actualError != tc.ExpectedError {
			t.Errorf("%s failed: expected %v error, got %v", tc.Description, tc.ExpectedError, actualError)
			continue
		}
		if actualResult != nil && actualResult.RoleArn != tc.ExpectedResult.RoleArn {
			t.Errorf("%s failed: expected %v result, got %v", tc.Description, tc.ExpectedResult, actualResult)
		}
	}
}

func TestCredentialCache_GetDefault(t *testing.T) {
	cases := []struct {
		CacheContents  map[string]*creds.RefreshableProvider
		DefaultRole    string
		Description    string
		ExpectedResult *creds.RefreshableProvider
		ExpectedError  error
	}{
		{
			Description:    "default role not in cache",
			DefaultRole:    "a",
			CacheContents:  make(map[string]*creds.RefreshableProvider),
			ExpectedError:  errors.NoCredentialsFoundInCache,
			ExpectedResult: nil,
		},
		{
			Description: "default role in cache",
			DefaultRole: "a",
			CacheContents: map[string]*creds.RefreshableProvider{
				"a": {RoleName: "a"},
			},
			ExpectedError:  nil,
			ExpectedResult: &creds.RefreshableProvider{RoleName: "a"},
		},
		{
			Description: "no default role set",
			DefaultRole: "",
			CacheContents: map[string]*creds.RefreshableProvider{
				"a": {RoleName: "a"},
			},
			ExpectedError:  errors.NoDefaultRoleSet,
			ExpectedResult: nil,
		},
		{
			Description: "default role in cache with assume",
			DefaultRole: "a",
			CacheContents: map[string]*creds.RefreshableProvider{
				"a":     {RoleName: "a"},
				"a/b/c": {RoleName: "a/b/c"},
			},
			ExpectedError:  nil,
			ExpectedResult: &creds.RefreshableProvider{RoleName: "a"},
		},
		{
			Description: "default assume role in cache",
			DefaultRole: "a/b/c",
			CacheContents: map[string]*creds.RefreshableProvider{
				"a/b/c": {RoleName: "a/b/c"},
			},
			ExpectedError:  nil,
			ExpectedResult: &creds.RefreshableProvider{RoleName: "a/b/c"},
		},
		{
			Description: "default assume role in cache with non-assume",
			DefaultRole: "a/b/c",
			CacheContents: map[string]*creds.RefreshableProvider{
				"a":     {RoleName: "a"},
				"a/b/c": {RoleName: "a/b/c"},
			},
			ExpectedError:  nil,
			ExpectedResult: &creds.RefreshableProvider{RoleName: "a/b/c"},
		},
	}

	for i, tc := range cases {
		t.Logf("test case %d: %s", i, tc.Description)
		testCache := CredentialCache{
			RoleCredentials: tc.CacheContents,
			DefaultRole:     tc.DefaultRole,
		}
		actualResult, actualError := testCache.GetDefault()
		if actualError != tc.ExpectedError {
			t.Errorf("%s failed: expected %v error, got %v", tc.Description, tc.ExpectedError, actualError)
			continue
		}
		if actualResult != nil && actualResult.RoleArn != tc.ExpectedResult.RoleArn {
			t.Errorf("%s failed: expected %v result, got %v", tc.Description, tc.ExpectedResult, actualResult)
		}
	}
}

func TestCredentialCache_SetDefault(t *testing.T) {
	testCache := CredentialCache{
		RoleCredentials: map[string]*creds.RefreshableProvider{},
	}
	expectedRole := "a"
	expectedExpiration := time.Unix(1, 0).Round(0)
	testClient, err := creds.GetTestClient(creds.ConsolemeCredentialResponseType{
		Credentials: &creds.AwsCredentials{
			AccessKeyId:     "a",
			SecretAccessKey: "b",
			SessionToken:    "c",
			Expiration:      creds.Time(time.Unix(1, 0)),
			RoleArn:         "e",
		},
	})
	if err != nil {
		t.Errorf("test setup failure: %e", err)
	}
	err = testCache.SetDefault(testClient, expectedRole, "b", make([]string, 0))
	if err != nil {
		t.Errorf("test failure: %e", err)
	}
	if testCache.DefaultRole != "a" {
		t.Errorf("got wrong default: expected %s, got %s", expectedRole, testCache.DefaultRole)
	}
	if testCache.RoleCredentials["a"].Expiration.Unix() != expectedExpiration.Unix() {
		t.Errorf("got wrong expiration: expected %s, got %s", expectedExpiration, testCache.RoleCredentials["a"].Expiration)
	}
}

func TestCredentialCache_DefaultLastUpdated(t *testing.T) {
	testCache := CredentialCache{
		RoleCredentials: map[string]*creds.RefreshableProvider{},
	}
	testClient, err := creds.GetTestClient(creds.ConsolemeCredentialResponseType{
		Credentials: &creds.AwsCredentials{
			AccessKeyId:     "a",
			SecretAccessKey: "b",
			SessionToken:    "c",
			Expiration:      creds.Time(time.Unix(1, 0)),
			RoleArn:         "e",
		},
	})
	if err != nil {
		t.Errorf("test setup failure: %e", err)
	}
	err = testCache.SetDefault(testClient, "a", "b", make([]string, 0))
	if err != nil {
		t.Errorf("test failure: %e", err)
	}
	timeFormat := "2006-01-02T15:04:05Z"
	result := testCache.DefaultLastUpdated()
	resultTime, err := time.Parse(timeFormat, result)
	if err != nil {
		t.Errorf("invalid time format returned: expected format %s, got result %s", timeFormat, result)
	}
	now := time.Now().Format(timeFormat)
	timeDiff := time.Now().Sub(resultTime)
	if timeDiff < 0*time.Second || timeDiff > 1*time.Second {
		t.Errorf("last refreshed time more than 1 second different than current time: %s, current time %s, difference %v seconds", result, now, timeDiff)
	}
}

func TestCredentialCache_DefaultLastUpdated_NoDefault(t *testing.T) {
	testCache := CredentialCache{
		RoleCredentials: map[string]*creds.RefreshableProvider{},
	}
	result := testCache.DefaultLastUpdated()
	if result != "" {
		t.Errorf("wrong last updated returned: got %s, expected empty string", result)
	}
}

func TestCredentialCache_DefaultArn(t *testing.T) {
	testCache := CredentialCache{
		RoleCredentials: map[string]*creds.RefreshableProvider{},
	}
	testClient, err := creds.GetTestClient(creds.ConsolemeCredentialResponseType{
		Credentials: &creds.AwsCredentials{
			AccessKeyId:     "a",
			SecretAccessKey: "b",
			SessionToken:    "c",
			Expiration:      creds.Time(time.Unix(1, 0)),
			RoleArn:         "e",
		},
	})
	if err != nil {
		t.Errorf("test setup failure: %e", err)
	}
	err = testCache.SetDefault(testClient, "a", "b", make([]string, 0))
	if err != nil {
		t.Errorf("test failure: %e", err)
	}
	expected := "e"
	result := testCache.DefaultArn()
	if result != expected {
		t.Errorf("wrong arn returned: got %s, expected %s", result, expected)
	}
}

func TestCredentialCache_DefaultArn_NoDefault(t *testing.T) {
	testCache := CredentialCache{
		RoleCredentials: map[string]*creds.RefreshableProvider{},
	}
	result := testCache.DefaultArn()
	if result != "" {
		t.Errorf("wrong arn returned: got %s, expected empty string", result)
	}
}

func TestCredentialCache_GetOrSet(t *testing.T) {
	cases := []struct {
		CacheContents  map[string]*creds.RefreshableProvider
		ClientResponse interface{}
		SearchString   string
		AssumeChain    []string
		Region         string
		Description    string
		ExpectedResult *creds.RefreshableProvider
		ExpectedError  error
	}{
		{
			Description:    "role not in cache",
			CacheContents:  make(map[string]*creds.RefreshableProvider),
			SearchString:   "a",
			AssumeChain:    []string{},
			ExpectedError:  nil,
			ExpectedResult: &creds.RefreshableProvider{RoleArn: "arn:aws:iam::012345678901:role/coolRole1"},
		},
		{
			Description: "role not in cache with assume",
			CacheContents: map[string]*creds.RefreshableProvider{
				"a/b/c": {RoleName: "a/b/c"},
			},
			SearchString:   "a",
			AssumeChain:    []string{},
			ExpectedError:  nil,
			ExpectedResult: &creds.RefreshableProvider{RoleArn: "arn:aws:iam::012345678901:role/coolRole2"},
		},
		{
			Description: "role already in cache",
			CacheContents: map[string]*creds.RefreshableProvider{
				"a": {RoleArn: "arn:aws:iam::012345678901:role/coolRole3"},
			},
			SearchString:   "a",
			AssumeChain:    []string{},
			ExpectedError:  nil,
			ExpectedResult: &creds.RefreshableProvider{RoleArn: "arn:aws:iam::012345678901:role/coolRole3"},
		},
	}

	for i, tc := range cases {
		t.Logf("test case %d: %s", i, tc.Description)
		testCache := CredentialCache{
			RoleCredentials: tc.CacheContents,
		}
		client, err := creds.GetTestClient(creds.ConsolemeCredentialResponseType{
			Credentials: &creds.AwsCredentials{
				AccessKeyId:     "a",
				SecretAccessKey: "b",
				SessionToken:    "c",
				Expiration:      creds.Time(time.Unix(1, 0)),
				RoleArn:         tc.ExpectedResult.RoleArn,
			},
		})
		if err != nil {
			t.Errorf("test setup failure: %e", err)
			continue
		}
		result, actualError := testCache.GetOrSet(client, tc.SearchString, tc.Region, tc.AssumeChain)
		if actualError != tc.ExpectedError {
			t.Errorf("%s failed: expected %v error, got %v", tc.Description, tc.ExpectedError, actualError)
			continue
		}
		if result == nil && tc.ExpectedResult != nil {
			t.Errorf("%s failed: got nil result, expected %v", tc.Description, tc.ExpectedResult)
			continue
		}
		if result != nil && result.RoleArn != tc.ExpectedResult.RoleArn {
			t.Errorf("%s failed: expected role %v, got %v", tc.Description, tc.ExpectedResult.RoleArn, result.RoleArn)
			continue
		}
	}
}
