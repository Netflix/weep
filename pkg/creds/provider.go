package creds

import (
	"context"
	"fmt"

	"github.com/netflix/weep/pkg/types"

	"github.com/netflix/weep/pkg/aws"
	"github.com/netflix/weep/pkg/creds/consoleme"
)

type IWeepCredentialProvider interface {
	CloseIdleConnections(context.Context)
	Credentials(context.Context, string, bool) (*aws.Credentials, error)
	List(context.Context) ([]string, error)
	ListExtended(context.Context) ([]types.RoleDetails, error)
	ResourceURL(context.Context, string) (string, error)
}

type IWeepRefreshableCredentialProvider interface {
	IWeepCredentialProvider
	AutoRefresh(context.Context)
}

var currentProvider IWeepCredentialProvider

func ensureProvider() {
	if currentProvider != nil {
		return
	}
	currentProvider = consoleme.NewProvider()
}

func Get(ctx context.Context, searchString string, ipRestrict bool, assumeChain []string) (*aws.Credentials, error) {
	ensureProvider()
	tempCreds, err := currentProvider.Credentials(ctx, searchString, ipRestrict)

	for _, assumeRoleArn := range assumeChain {
		tempCreds.AccessKeyId, tempCreds.SecretAccessKey, tempCreds.SessionToken, err = aws.GetAssumeRoleCredentials(
			tempCreds.AccessKeyId, tempCreds.SecretAccessKey, tempCreds.SessionToken, assumeRoleArn)
		if err != nil {
			return nil, fmt.Errorf("role assumption failed for %s: %s", assumeRoleArn, err)
		}
	}

	return tempCreds, nil
}

func List(ctx context.Context) ([]string, error) {
	ensureProvider()
	return currentProvider.List(ctx)
}

func ListExtended(ctx context.Context) ([]types.RoleDetails, error) {
	ensureProvider()
	return currentProvider.ListExtended(ctx)
}

func ResourceURL(ctx context.Context, arn string) (string, error) {
	ensureProvider()
	return currentProvider.ResourceURL(ctx, arn)
}
