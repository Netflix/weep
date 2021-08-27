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

	"github.com/spf13/viper"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

// getSessionName returns the AWS session name, or defaults to weep if we can't find one.
func getSessionName(session *sts.STS) string {
	identity, err := session.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		log.Warnf("could not get user identity; defaulting to weep: %s", err)
		return "weep"
	}

	// split identity.UserId on colon, which should give us a 2-element slice with the principal ID and session name
	splitId := strings.Split(*identity.UserId, ":")
	if len(splitId) < 2 {
		log.Warnf("session name not found; defaulting to weep")
		return "weep"
	}

	return splitId[1]
}

// getAssumeRoleCredentials uses the provided credentials to assume the role specified by roleArn.
func getAssumeRoleCredentials(id, secret, token, roleArn string) (string, string, string, error) {
	region := viper.GetString("aws.region")
	staticCreds := credentials.NewStaticCredentials(id, secret, token)
	awsSession := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Credentials: staticCreds,
			Region:      aws.String(region),
		},
	}))

	stsSession := sts.New(awsSession)
	sessionName := getSessionName(stsSession)

	stsParams := &sts.AssumeRoleInput{
		RoleArn:         &roleArn,
		RoleSessionName: &sessionName,
		DurationSeconds: aws.Int64(3600),
	}

	stsCreds, err := stsSession.AssumeRole(stsParams)
	if err != nil {
		return "", "", "", fmt.Errorf("error retrieving awsSession token: %s", err)
	}
	return *stsCreds.Credentials.AccessKeyId, *stsCreds.Credentials.SecretAccessKey, *stsCreds.Credentials.SessionToken, nil
}

// GetCredentialsC uses the provided Client to request credentials from ConsoleMe then
// follows the provided chain of roles to assume. Roles are assumed in the order in which
// they appear in the assumeRole slice.
func GetCredentialsC(client HTTPClient, role string, ipRestrict bool, assumeRole []string) (*AwsCredentials, error) {
	resp, err := client.GetRoleCredentials(role, ipRestrict)
	if err != nil {
		return nil, err
	}

	for _, assumeRoleArn := range assumeRole {
		resp.AccessKeyId, resp.SecretAccessKey, resp.SessionToken, err = getAssumeRoleCredentials(resp.AccessKeyId, resp.SecretAccessKey, resp.SessionToken, assumeRoleArn)
		if err != nil {
			return nil, fmt.Errorf("role assumption failed for %s: %s", assumeRoleArn, err)
		}
	}

	return resp, nil
}

// GetCredentials requests credentials from ConsoleMe then follows the provided chain of roles to
// assume. Roles are assumed in the order in which they appear in the assumeRole slice.
func GetCredentials(role string, ipRestrict bool, assumeRole []string, region string) (*AwsCredentials, error) {
	client, err := GetClient(region)
	if err != nil {
		return nil, err
	}

	return GetCredentialsC(client, role, ipRestrict, assumeRole)
}
