package metadata

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/netflix/weep/util"
	log "github.com/sirupsen/logrus"

	"github.com/netflix/weep/consoleme"
)

var (
	Role                string
	NoIpRestrict        bool
	MetaDataCredentials consoleme.AwsCredentials
	MetadataRegion      string
	LastRenewal         time.Time
)

func StartMetaDataRefresh(client *consoleme.Client) {
	retryDelay := 5 * time.Second
	retryCount := 10
	var err error
	for {
		// TODO: If 403 response,
		MetaDataCredentials, err = client.GetRoleCredentials(Role, NoIpRestrict)
		util.CheckError(err)
		sess, err := session.NewSession(&aws.Config{
			Credentials: credentials.NewStaticCredentials(
				MetaDataCredentials.AccessKeyId,
				MetaDataCredentials.SecretAccessKey,
				MetaDataCredentials.SessionToken),
		})
		util.CheckError(err)
		svc := sts.New(sess)
		input := &sts.GetCallerIdentityInput{}

		result, err := svc.GetCallerIdentity(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				default:
					fmt.Println(aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				fmt.Println(err.Error())
			}
			return
		}
		// Replace assumed role ARN with role ARN, if possible
		// arn:aws:sts::123456789012:assumed-role/exampleInstanceProfile/user@example.com ->
		// arn:aws:iam::123456789012:role/exampleInstanceProfile
		Role = strings.Replace(*result.Arn, ":sts:", ":iam:", 1)
		Role = strings.Replace(Role, ":assumed-role/", ":role/", 1)
		// result.UserId looks like AROAIEBAVBLAH:user@example.com
		splittedUserId := strings.Split(*result.UserId, ":")
		if len(splittedUserId) > 1 {
			sessionName := splittedUserId[1]
			Role = strings.Replace(
				Role,
				fmt.Sprintf("/%s", sessionName),
				"",
				1)
		}
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
