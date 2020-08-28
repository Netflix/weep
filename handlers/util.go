package handlers

import "strings"

var (
	whitelist [7]string
)

func init() {
	whitelist[0] = "aws-sdk-"
	whitelist[1] = "Botocore/"
	whitelist[2] = "Boto3/"
	whitelist[3] = "aws-cli/"
	whitelist[4] = "aws-chalice/"
	whitelist[5] = "nflx-"
	whitelist[6] = "eureka-java-client"
}

func checkUserAgent(ua string) bool {
	for i := range whitelist {
		if strings.HasPrefix(ua, whitelist[i]) {
			return true
		}
	}
	return false
}
