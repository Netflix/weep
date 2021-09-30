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

package aws

import (
	"fmt"
	"strings"

	"github.com/netflix/weep/pkg/logging"

	"github.com/spf13/viper"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
)

// getSessionName returns the AWS session name, or defaults to weep if we can't find one.
func getSessionName(session *sts.STS) string {
	identity, err := session.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		logging.Log.Warnf("could not get user identity; defaulting to weep: %s", err)
		return "weep"
	}

	// split identity.UserId on colon, which should give us a 2-element slice with the principal ID and session name
	splitId := strings.Split(*identity.UserId, ":")
	if len(splitId) < 2 {
		logging.Log.Warnf("session name not found; defaulting to weep")
		return "weep"
	}

	return splitId[1]
}

// GetAssumeRoleCredentials uses the provided credentials to assume the role specified by roleArn.
func GetAssumeRoleCredentials(id, secret, token, roleArn string) (string, string, string, error) {
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

func GetSession() *session.Session {
	return session.Must(session.NewSession())
}

func GetCallerIdentity(awsSession *session.Session) (*sts.GetCallerIdentityOutput, error) {
	if awsSession == nil {
		awsSession = GetSession()
	}
	stsSession := sts.New(awsSession)
	input := &sts.GetCallerIdentityInput{}
	return stsSession.GetCallerIdentity(input)
}

func ListAccountAliases(awsSession *session.Session) ([]*string, error) {
	aliases := make([]*string, 0)
	pageNum := 0
	if awsSession == nil {
		awsSession = GetSession()
	}
	iamSession := iam.New(awsSession)
	input := &iam.ListAccountAliasesInput{}
	err := iamSession.ListAccountAliasesPages(input, func(page *iam.ListAccountAliasesOutput, lastPage bool) bool {
		pageNum++
		fmt.Println(page)
		aliases = append(aliases, page.AccountAliases...)
		return !*page.IsTruncated
	})
	if err != nil {
		return nil, err
	}
	return aliases, nil
}
