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

	"github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/netflix/weep/errors"
)

var (
	testAccessKeyId     = "a"
	testSecretAccessKey = "b"
	testSessionToken    = "c"
	testRegion          = "d"
	testRole            = "e"
	testRoleArn         = "f"
	testProviderName    = "g"
	testExpiration      = Time(time.Now().Add(60 * time.Minute).Round(0))
	testSoonExpiration  = Time(time.Now().Add(5 * time.Minute).Round(0))
	testPastExpiration  = Time(time.Now().Add(-5 * time.Minute).Round(0))
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
				LastRefreshed: Time{},
				Region:        testRegion,
				RoleName:      testRole,
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
				LastRefreshed: Time{},
				Region:        testRegion,
				RoleName:      testRole,
				RoleArn:       testRoleArn,
				NoIpRestrict:  true,
				AssumeChain:   make([]string, 0),
			},
		},
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
		if actualResult != nil && actualResult.Expiration.Unix() != tc.ExpectedResult.Expiration.Unix() {
			t.Errorf("%s failed: got %v expiration, expected %v", tc.Description, actualResult.Expiration, tc.ExpectedResult.Expiration)
			continue
		}
		if actualResult != nil && actualResult.Region != tc.ExpectedResult.Region {
			t.Errorf("%s failed: got %v region, expected %v", tc.Description, actualResult.Region, tc.ExpectedResult.Region)
			continue
		}
		if actualResult != nil && actualResult.RoleName != tc.ExpectedResult.RoleName {
			t.Errorf("%s failed: got %v role, expected %v", tc.Description, actualResult.RoleName, tc.ExpectedResult.RoleName)
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
			Retries:      2,
			RetryDelay:   1,
			CredentialResponse: ConsolemeCredentialErrorMessageType{
				Code:    "901",
				Message: "Nope",
			},
			ExpectedError:  errors.MultipleMatchingRoles,
			ExpectedResult: nil,
		},
	}

	zeroTime := Time{}
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
			RoleName:     tc.Role,
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

func TestRefreshableProvider_checkAndRefresh(t *testing.T) {
	cases := []struct {
		Description        string
		Expiration         Time
		ExpectedExpiration Time
		CredentialResponse interface{}
		ShouldRefresh      bool
		ExpectedError      error
	}{
		{
			Description:        "not ready to refresh",
			Expiration:         testExpiration,
			ExpectedExpiration: testExpiration,
			CredentialResponse: testCredentialResponse,
			ShouldRefresh:      false,
			ExpectedError:      nil,
		},
		{
			Description:        "ready to refresh",
			Expiration:         testSoonExpiration,
			ExpectedExpiration: testExpiration,
			CredentialResponse: testCredentialResponse,
			ShouldRefresh:      true,
			ExpectedError:      nil,
		},
		{
			Description:        "ready to refresh",
			Expiration:         testPastExpiration,
			ExpectedExpiration: testExpiration,
			CredentialResponse: testCredentialResponse,
			ShouldRefresh:      true,
			ExpectedError:      nil,
		},
	}

	for i, tc := range cases {
		t.Logf("test case %d: %s", i, tc.Description)
		client, err := GetTestClient(tc.CredentialResponse)
		if err != nil {
			t.Errorf("test setup failure: %e", err)
			continue
		}
		rp := RefreshableProvider{
			client:     client,
			Expiration: tc.Expiration,
			retries:    1,
		}
		refreshed, err := rp.checkAndRefresh(10)
		if err != tc.ExpectedError {
			t.Errorf("%s failed: expected %v error, got %v", tc.Description, tc.ExpectedError, err)
		} else {
			continue
		}
		if refreshed != tc.ShouldRefresh {
			t.Errorf("%s failed: expected %v, got %v", tc.Description, tc.ShouldRefresh, refreshed)
		}
		if rp.Expiration != tc.Expiration {
			t.Errorf("%s failed: expected expiration %v, got %v", tc.Description, tc.Expiration, rp.Expiration)
		}
	}
}

func TestRefreshableProvider_IsExpired(t *testing.T) {
	t.Logf("test case: check IsExpired is always false")
	client, err := GetTestClient(testCredentialResponse)
	if err != nil {
		t.Errorf("test setup failure: %e", err)
		t.Fail()
	}

	rp := RefreshableProvider{
		client: client,
	}

	if rp.IsExpired() {
		t.Errorf("failed: IsExpired returned true")
	}
}

func TestRefreshableProvider_Retrieve(t *testing.T) {
	t.Logf("test case: retrieve credentials")

	expected := credentials.Value{
		AccessKeyID:     testAccessKeyId,
		SecretAccessKey: testSecretAccessKey,
		SessionToken:    testSessionToken,
		ProviderName:    testProviderName,
	}

	rp := RefreshableProvider{
		value: expected,
	}

	result, err := rp.Retrieve()
	if err != nil {
		t.Errorf("failed: expected nil error, got %v", err)
	}
	if result != expected {
		t.Errorf("failed: expected %v, got %v", expected, result)
	}
}
