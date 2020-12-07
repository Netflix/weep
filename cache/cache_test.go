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
				"a": {Role: "a"},
			},
			Role:           "a",
			AssumeChain:    []string{},
			ExpectedError:  nil,
			ExpectedResult: &creds.RefreshableProvider{Role: "a"},
		},
		{
			Description: "role in cache with assume",
			CacheContents: map[string]*creds.RefreshableProvider{
				"a":     {Role: "a"},
				"a/b/c": {Role: "a/b/c"},
			},
			Role:           "a",
			AssumeChain:    []string{},
			ExpectedError:  nil,
			ExpectedResult: &creds.RefreshableProvider{Role: "a"},
		},
		{
			Description: "assume role in cache",
			CacheContents: map[string]*creds.RefreshableProvider{
				"a/b/c": {Role: "a/b/c"},
			},
			Role:           "a",
			AssumeChain:    []string{"b", "c"},
			ExpectedError:  nil,
			ExpectedResult: &creds.RefreshableProvider{Role: "a/b/c"},
		},
		{
			Description: "assume role in cache with non-assume",
			CacheContents: map[string]*creds.RefreshableProvider{
				"a":     {Role: "a"},
				"a/b/c": {Role: "a/b/c"},
			},
			Role:           "a",
			AssumeChain:    []string{"b", "c"},
			ExpectedError:  nil,
			ExpectedResult: &creds.RefreshableProvider{Role: "a/b/c"},
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
		if actualResult != nil && actualResult.Role != tc.ExpectedResult.Role {
			t.Errorf("%s failed: expected %v result, got %v", tc.Description, tc.ExpectedResult, actualResult)
		}
	}
}
