package v2

import (
	"context"
	"github.com/netflix/weep/pkg/aws"
)

type IWeepCredentialProvider interface {
	Credentials(context.Context, string, bool) (*aws.Credentials, error)
	List(context.Context) ([]string, error)
	ListExtended(context.Context) ([]RoleDetails, error)
	ResourceURL(context.Context, string) (string, error)
}

type IWeepRefreshableCredentialProvider interface {
	IWeepCredentialProvider
	AutoRefresh(context.Context)
}

// RoleDetails represents the response structure of ConsoleMe's model for detailed eligible roles
type RoleDetails struct {
	Arn           string `json:"arn"`
	AccountNumber string `json:"account_id"`
	AccountName   string `json:"account_friendly_name"`
	RoleName      string `json:"role_name"`
	Apps          struct {
		AppDetails []AppDetails `json:"app_details"`
	} `json:"apps"`
}

// AppDetails represents the structure of details returned by ConsoleMe about a single app
type AppDetails struct {
	Name     string `json:"name"`
	Owner    string `json:"owner"`
	OwnerURL string `json:"owner_url"`
	AppURL   string `json:"app_url"`
}
