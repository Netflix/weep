package consoleme

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/netflix/weep/pkg/types"

	"github.com/netflix/weep/pkg/aws"
	"github.com/netflix/weep/pkg/creds/consoleme/challenge"
	"github.com/netflix/weep/pkg/creds/consoleme/mtls"
	"github.com/spf13/viper"
)

type CredentialProvider struct {
	client       *http.Client
	forceRefresh bool
}

func NewProvider() *CredentialProvider {
	return &CredentialProvider{}
}

func defaultTransport() *http.Transport {
	timeout := time.Duration(viper.GetInt("server.http_timeout")) * time.Second
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
	}
}

func (c *CredentialProvider) ensureClient() error {
	if !c.forceRefresh && c.client != nil {
		return nil
	}
	var client *http.Client
	var err error
	consoleMeUrl := viper.GetString("consoleme_url")
	authenticationMethod := viper.GetString("authentication_method")

	if authenticationMethod == "mtls" {
		client, err = mtls.NewHTTPClient()
		if err != nil {
			return err
		}
	} else if authenticationMethod == "challenge" {
		err := challenge.RefreshChallenge()
		if err != nil {
			return err
		}
		client, err = challenge.NewHTTPClient(consoleMeUrl)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Authentication method unsupported or not provided.")
	}
	client.Transport = defaultTransport()
	c.client = client
	return nil
}

func (c *CredentialProvider) Credentials(ctx context.Context, searchString string, ipRestrict bool) (*aws.Credentials, error) {
	if err := c.ensureClient(); err != nil {
		return nil, err
	}
	return retrieveCredentials(ctx, c.client, searchString, ipRestrict)
}

func (c *CredentialProvider) List(ctx context.Context) ([]string, error) {
	if err := c.ensureClient(); err != nil {
		return nil, err
	}
	return retrieveRoles(ctx, c.client)
}

func (c *CredentialProvider) ListExtended(ctx context.Context) ([]types.RoleDetails, error) {
	if err := c.ensureClient(); err != nil {
		return nil, err
	}
	return retrieveRolesExtended(ctx, c.client)
}

func (c *CredentialProvider) ResourceURL(ctx context.Context, arn string) (string, error) {
	if err := c.ensureClient(); err != nil {
		return "", err
	}
	return retrieveResourceURL(ctx, c.client, arn)
}

func (c *CredentialProvider) CloseIdleConnections(ctx context.Context) {
	c.client.CloseIdleConnections()
}
