package v2

import (
	"github.com/netflix/weep/pkg/aws"
	"io"
	"net/http"
)

type IWeepClient interface {
	GetRoleCredentials(role string, ipRestrict bool) (*aws.Credentials, error)
	CloseIdleConnections()
	buildRequest(string, string, io.Reader, string) (*http.Request, error)
}