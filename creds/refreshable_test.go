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
	"testing"
	"time"

	"github.com/netflix/weep/errors"
)

var (
	testAccessKeyId     = "a"
	testSecretAccessKey = "b"
	testSessionToken    = "c"
	testRegion          = "d"
	testRole            = "e"
	testRoleArn         = "f"
	testExpiration      = time.Unix(1, 0)
	testCredentials     = &AwsCredentials{
		AccessKeyId:     testAccessKeyId,
		SecretAccessKey: testSecretAccessKey,
		SessionToken:    testSessionToken,
		Expiration:      testExpiration,
		RoleArn:         testRoleArn,
	}
	testCredentialResponse = ConsolemeCredentialResponseType{
		Credentials: testCredentials,
	}
)

func TestNewRefreshableProvider(t *testing.T) {
	cases := []struct {
		Description        string
		Role               string
		Region             string
		AssumeChain        []string
		NoIpRestrict       bool
		CredentialResponse interface{}
		ExpectedError      error
		ExpectedResult     *RefreshableProvider
	}{
		{
			Description:        "create refreshable provider with IP restrict",
			Role:               testRole,
			Region:             testRegion,
			AssumeChain:        make([]string, 0),
			NoIpRestrict:       false,
			CredentialResponse: testCredentialResponse,
			ExpectedError:      nil,
			ExpectedResult: &RefreshableProvider{
				Expiration:    testExpiration,
				LastRefreshed: time.Time{},
				Region:        testRegion,
				Role:          testRole,
				RoleArn:       testRoleArn,
				NoIpRestrict:  false,
				AssumeChain:   make([]string, 0),
			},
		},
		{
			Description:        "create refreshable provider with no IP restrict",
			Role:               testRole,
			Region:             testRegion,
			AssumeChain:        make([]string, 0),
			NoIpRestrict:       true,
			CredentialResponse: testCredentialResponse,
			ExpectedError:      nil,
			ExpectedResult: &RefreshableProvider{
				Expiration:    testExpiration,
				LastRefreshed: time.Time{},
				Region:        testRegion,
				Role:          testRole,
				RoleArn:       testRoleArn,
				NoIpRestrict:  true,
				AssumeChain:   make([]string, 0),
			},
		},
		//{
		//	Description:  "bad credential response to check error handling",
		//	Role:         testRole,
		//	Region:       testRegion,
		//	AssumeChain:  make([]string, 0),
		//	NoIpRestrict: true,
		//	CredentialResponse: ConsolemeCredentialErrorMessageType{
		//		Code:    "403",
		//		Message: "Nope",
		//	},
		//	ExpectedError:  errors.CredentialRetrievalError,
		//	ExpectedResult: nil,
		//},
	}

	for i, tc := range cases {
		t.Logf("test case %d: %s", i, tc.Description)
		client, err := GetTestClient(tc.CredentialResponse)
		if err != nil {
			t.Errorf("test setup failure: %e", err)
			continue
		}
		actualResult, actualError := NewRefreshableProvider(client, tc.Role, tc.Region, tc.AssumeChain, tc.NoIpRestrict)
		if err != tc.ExpectedError {
			t.Errorf("%s failed: expected %v error, got %v", tc.Description, tc.ExpectedError, actualError)
		}
		if actualResult == nil && tc.ExpectedResult != nil {
			t.Errorf("%s failed: got nil result, expected %v", tc.Description, tc.ExpectedResult)
			continue
		}
		if actualResult != nil && actualResult.Expiration != tc.ExpectedResult.Expiration {
			t.Errorf("%s failed: got %v expiration, expected %v", tc.Description, actualResult.Expiration, tc.ExpectedResult.Expiration)
			continue
		}
		if actualResult != nil && actualResult.Region != tc.ExpectedResult.Region {
			t.Errorf("%s failed: got %v region, expected %v", tc.Description, actualResult.Region, tc.ExpectedResult.Region)
			continue
		}
		if actualResult != nil && actualResult.Role != tc.ExpectedResult.Role {
			t.Errorf("%s failed: got %v role, expected %v", tc.Description, actualResult.Role, tc.ExpectedResult.Role)
			continue
		}
		if actualResult != nil && actualResult.RoleArn != tc.ExpectedResult.RoleArn {
			t.Errorf("%s failed: got %v role ARN, expected %v", tc.Description, actualResult.RoleArn, tc.ExpectedResult.RoleArn)
			continue
		}
		if actualResult != nil && actualResult.NoIpRestrict != tc.ExpectedResult.NoIpRestrict {
			t.Errorf("%s failed: got %v region, expected %v", tc.Description, actualResult.NoIpRestrict, tc.ExpectedResult.NoIpRestrict)
			continue
		}
	}

}

func TestRefreshableProvider_refresh(t *testing.T) {
	cases := []struct {
		Description        string
		Role               string
		RoleArn            string
		Region             string
		AssumeChain        []string
		NoIpRestrict       bool
		Retries            int
		RetryDelay         int
		CredentialResponse interface{}
		ExpectedError      error
		ExpectedResult     *RefreshableProvider
	}{
		{
			Description:        "happy path refresh",
			Role:               testRole,
			RoleArn:            testRoleArn,
			Region:             testRegion,
			AssumeChain:        make([]string, 0),
			NoIpRestrict:       false,
			Retries:            5,
			RetryDelay:         5,
			CredentialResponse: testCredentialResponse,
			ExpectedError:      nil,
			ExpectedResult:     &RefreshableProvider{},
		},
		{
			Description:  "bad credential response",
			Role:         testRole,
			RoleArn:      testRoleArn,
			Region:       testRegion,
			AssumeChain:  make([]string, 0),
			NoIpRestrict: false,
			Retries:      1,
			RetryDelay:   1,
			CredentialResponse: ConsolemeCredentialErrorMessageType{
				Code:    "403",
				Message: "Nope",
			},
			ExpectedError:  errors.CredentialRetrievalError,
			ExpectedResult: nil,
		},
	}

	zeroTime := time.Time{}
	for i, tc := range cases {
		t.Logf("test case %d: %s", i, tc.Description)
		client, err := GetTestClient(tc.CredentialResponse)
		if err != nil {
			t.Errorf("test setup failure: %e", err)
			continue
		}
		rp := RefreshableProvider{
			client:       client,
			retries:      tc.Retries,
			retryDelay:   tc.RetryDelay,
			Region:       tc.Region,
			Role:         tc.Role,
			RoleArn:      tc.RoleArn,
			NoIpRestrict: tc.NoIpRestrict,
			AssumeChain:  tc.AssumeChain,
		}
		// pre-refresh checks
		if rp.value.SessionToken != "" || rp.value.AccessKeyID != "" || rp.value.SecretAccessKey != "" || rp.value.ProviderName != "" {
			t.Errorf("%s failed: credential values should not exist: %v", tc.Description, rp.value)
			continue
		}
		if rp.Expiration != zeroTime {
			t.Errorf("%s failed: expiration should not be set, got %v", tc.Description, rp.Expiration)
			continue
		}
		// perform refresh
		err = rp.refresh()
		// post-refresh checks
		if err != tc.ExpectedError {
			t.Errorf("%s failed: expected %v error, got %v", tc.Description, tc.ExpectedError, err)
		} else {
			continue
		}
		if rp.value.SessionToken == "" || rp.value.AccessKeyID == "" || rp.value.SecretAccessKey == "" || rp.value.ProviderName == "" {
			t.Errorf("%s failed: credential values should not be empty: %v", tc.Description, rp.value)
		}
		if rp.Expiration == zeroTime {
			t.Errorf("%s failed: Expiration should be set, got %v", tc.Description, rp.Expiration)
		}
		if rp.LastRefreshed == zeroTime {
			t.Errorf("%s failed: LastRefreshed should be set, got %v", tc.Description, rp.Expiration)
		}
	}
}
