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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

func getAwsCredentials(role string, ipRestrict bool) (string, string, string, error) {
	client, err := GetClient()
	if err != nil {
		return "", "", "", err
	}
	tempCreds, err := client.GetRoleCredentials(role, ipRestrict)
	if err != nil {
		return "", "", "", err
	}

	return tempCreds.AccessKeyId, tempCreds.SecretAccessKey, tempCreds.SessionToken, nil
}

func getAssumeRoleCredentials(id, secret, token, roleArn string) (string, string, string, error) {
	staticCreds := credentials.NewStaticCredentials(id, secret, token)
	awsSession := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Credentials: staticCreds,
		},
	}))

	stsSession := sts.New(awsSession)

	stsParams := &sts.AssumeRoleInput{
		RoleArn:         &roleArn,
		RoleSessionName: aws.String("weep"),
		DurationSeconds: aws.Int64(3600),
	}

	stsCreds, err := stsSession.AssumeRole(stsParams)
	if err != nil {
		return "", "", "", fmt.Errorf("error retrieving awsSession token: %s", err)
	}
	return *stsCreds.Credentials.AccessKeyId, *stsCreds.Credentials.SecretAccessKey, *stsCreds.Credentials.SessionToken, nil
}

func GetCredentials(role string, ipRestrict bool, assumeRole ...string) (*AwsCredentials, error) {
	id, secret, token, err := getAwsCredentials(role, ipRestrict)
	if err != nil {
		return nil, err
	}

	for _, assumeRoleArn := range assumeRole {
		id, secret, token, err = getAssumeRoleCredentials(id, secret, token, assumeRoleArn)
		if err != nil {
			return nil, fmt.Errorf("role assumption failed for %s: %s", assumeRoleArn, err)
		}
	}

	finalCreds := &AwsCredentials{
		AccessKeyId:     id,
		SecretAccessKey: secret,
		SessionToken:    token,
	}
	return finalCreds, nil
}
