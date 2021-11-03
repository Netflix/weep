package creds

import (
	"io"
	"net/http"

	"github.com/netflix/weep/pkg/aws"
)

type IWeepClient interface {
	GetRoleCredentials(role string, ipRestrict bool) (*aws.Credentials, error)
	CloseIdleConnections()
	buildRequest(string, string, io.Reader, string) (*http.Request, error)
}
