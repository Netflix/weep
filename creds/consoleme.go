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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/netflix/weep/metadata"

	"github.com/netflix/weep/logging"

	werrors "github.com/netflix/weep/errors"
	"github.com/spf13/viper"

	"github.com/netflix/weep/httpAuth/challenge"
	"github.com/netflix/weep/httpAuth/mtls"

	"github.com/pkg/errors"
)

var log = logging.GetLogger()
var clientVersion = fmt.Sprintf("%s", metadata.Version)

var userAgent = "weep/" + clientVersion + " Go-http-client/1.1"

type Account struct {
}

// HTTPClient is the interface we expect HTTP clients to implement.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
	GetRoleCredentials(role string, ipRestrict bool) (*AwsCredentials, error)
	CloseIdleConnections()
	buildRequest(string, string, io.Reader, string) (*http.Request, error)
}

// Client represents a ConsoleMe client.
type Client struct {
	http.Client
	Host   string
	Region string
}

// GetClient creates an authenticated ConsoleMe client
func GetClient(region string) (*Client, error) {
	var client *Client
	consoleMeUrl := viper.GetString("consoleme_url")
	authenticationMethod := viper.GetString("authentication_method")

	if authenticationMethod == "mtls" {
		mtlsClient, err := mtls.NewHTTPClient()
		if err != nil {
			return client, err
		}
		client, err = NewClient(consoleMeUrl, "", mtlsClient)
		if err != nil {
			return client, err
		}
	} else if authenticationMethod == "challenge" {
		err := challenge.RefreshChallenge()
		if err != nil {
			return client, err
		}
		httpClient, err := challenge.NewHTTPClient(consoleMeUrl)
		if err != nil {
			return client, err
		}
		client, err = NewClient(consoleMeUrl, "", httpClient)
		if err != nil {
			return client, err
		}
	} else {
		return nil, fmt.Errorf("Authentication method unsupported or not provided.")
	}

	return client, nil
}

// NewClient takes a ConsoleMe hostname and *http.Client, and returns a
// ConsoleMe client that will talk to that ConsoleMe instance for AWS Credentials.
func NewClient(hostname string, region string, httpc *http.Client) (*Client, error) {
	if len(hostname) == 0 {
		return nil, errors.New("hostname cannot be empty string")
	}

	if httpc == nil {
		httpc = &http.Client{Transport: defaultTransport()}
	}

	c := &Client{
		Client: *httpc,
		Host:   hostname,
		Region: region,
	}

	return c, nil
}

func (c *Client) buildRequest(method string, resource string, body io.Reader, apiPrefix string) (*http.Request, error) {
	urlStr := c.Host + apiPrefix + resource
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Add("Content-Type", "application/json")

	return req, nil
}

// CloseIdleConnections calls CloseIdleConnections() on the client's HTTP transport.
func (c *Client) CloseIdleConnections() {
	transport, ok := c.Client.Transport.(*http.Transport)
	if !ok {
		// This is unlikely, but we'll fail out anyway.
		return
	}
	transport.CloseIdleConnections()
}

// accounts returns all accounts, and allows you to filter the accounts by sub-resources
// like: /accounts/service/support
func (c *Client) Roles() ([]string, error) {
	req, err := c.buildRequest(http.MethodGet, "/get_roles", nil, "/api/v1")
	if err != nil {
		return nil, errors.Wrap(err, "failed to build request")
	}

	// Add URL Parameters
	q := url.Values{}
	q.Add("all", "true")
	req.URL.RawQuery = q.Encode()

	resp, err := c.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to action request")
	}

	defer resp.Body.Close()
	document, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, parseError(resp.StatusCode, document)
	}

	var roles []string
	if err := json.Unmarshal(document, &roles); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal JSON")
	}

	return roles, nil
}

// Get resource URL given an ARN, from ConsoleMe
func (c *Client) GetResourceURL(arn string) (string, error) {
	req, err := c.buildRequest(http.MethodGet, "/get_resource_url", nil, "/api/v2")
	if err != nil {
		return "", errors.Wrap(err, "failed to build request")
	}

	// Add URL Parameters
	q := url.Values{}
	q.Add("arn", arn)
	req.URL.RawQuery = q.Encode()

	resp, err := c.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to action request")
	}

	defer resp.Body.Close()
	document, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read response body")
	}
	if resp.StatusCode != http.StatusOK {
		return "", parseWebError(document)
	}
	var responseParsed ConsolemeWebResponse
	if err := json.Unmarshal(document, &responseParsed); err != nil {
		return "", errors.Wrap(err, "failed to unmarshal JSON")
	}
	return baseWebURL() + responseParsed.Data["url"], nil
}

// baseWebURL allows the ConsoleMe URL to be overridden for cases where the API
// and UI are accessed via different URLs
func baseWebURL() string {
	if override := viper.GetString("consoleme_open_url_override"); override != "" {
		return override
	}
	return viper.GetString("consoleme_url")
}

func parseWebError(rawErrorResponse []byte) error {
	var errorResponse ConsolemeWebResponse
	if err := json.Unmarshal(rawErrorResponse, &errorResponse); err != nil {
		return errors.Wrap(err, "failed to unmarshal JSON")
	}
	return fmt.Errorf(strings.Join(errorResponse.Errors, "\n"))
}

func parseError(statusCode int, rawErrorResponse []byte) error {
	var errorResponse ConsolemeCredentialErrorMessageType
	if err := json.Unmarshal(rawErrorResponse, &errorResponse); err != nil {
		return errors.Wrap(err, "failed to unmarshal JSON")
	}

	switch errorResponse.Code {
	case "899":
		return werrors.InvalidArn
	case "900":
		return werrors.NoMatchingRoles
	case "901":
		return werrors.MultipleMatchingRoles
	case "902":
		return werrors.CredentialRetrievalError
	case "903":
		return werrors.NoMatchingRoles
	case "904":
		return werrors.MalformedRequestError
	case "905":
		return werrors.MutualTLSCertNeedsRefreshError
	case "invalid_jwt":
		log.Errorf("Authentication is invalid or has expired. Please restart weep to re-authenticate.")
		err := challenge.DeleteLocalWeepCredentials()
		if err != nil {
			log.Errorf("failed to delete credentials: %v", err)
		}
		return werrors.InvalidJWT
	default:
		return fmt.Errorf("unexpected HTTP status %d, want 200. Response: %s", statusCode, string(rawErrorResponse))
	}
}

func (c *Client) GetRoleCredentials(role string, ipRestrict bool) (*AwsCredentials, error) {
	return getRoleCredentialsFunc(c, role, ipRestrict)
}

func getRoleCredentialsFunc(c HTTPClient, role string, ipRestrict bool) (*AwsCredentials, error) {
	var credentialsResponse ConsolemeCredentialResponseType

	cmCredRequest := ConsolemeCredentialRequestType{
		RequestedRole:  role,
		NoIpRestricton: ipRestrict,
	}

	if metadataEnabled := viper.GetBool("feature_flags.consoleme_metadata"); metadataEnabled == true {
		cmCredRequest.Metadata = metadata.GetInstanceInfo()
	}

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(cmCredRequest)
	if err != nil {
		return credentialsResponse.Credentials, errors.Wrap(err, "failed to create request body")
	}

	req, err := c.buildRequest(http.MethodPost, "/get_credentials", b, "/api/v1")
	if err != nil {
		return credentialsResponse.Credentials, errors.Wrap(err, "failed to build request")
	}

	resp, err := c.Do(req)
	if err != nil {
		return credentialsResponse.Credentials, errors.Wrap(err, "failed to action request")
	}

	defer resp.Body.Close()
	document, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return credentialsResponse.Credentials, errors.Wrap(err, "failed to read response body")
	}
	if resp.StatusCode != http.StatusOK {
		return credentialsResponse.Credentials, parseError(resp.StatusCode, document)
	}

	if err := json.Unmarshal(document, &credentialsResponse); err != nil {
		return credentialsResponse.Credentials, errors.Wrap(err, "failed to unmarshal JSON")
	}

	if credentialsResponse.Credentials == nil {
		return nil, werrors.CredentialRetrievalError
	}

	return credentialsResponse.Credentials, nil
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

type ClientMock struct {
	DoFunc                 func(req *http.Request) (*http.Response, error)
	GetRoleCredentialsFunc func(role string, ipRestrict bool) (*AwsCredentials, error)
}

func (c *ClientMock) GetRoleCredentials(role string, ipRestrict bool) (*AwsCredentials, error) {
	return getRoleCredentialsFunc(c, role, ipRestrict)
}

func (c *ClientMock) CloseIdleConnections() {}

func (c *ClientMock) buildRequest(string, string, io.Reader, string) (*http.Request, error) {
	return &http.Request{}, nil
}

func (c *ClientMock) Do(req *http.Request) (*http.Response, error) {
	return c.DoFunc(req)
}

func GetTestClient(responseBody interface{}) (HTTPClient, error) {
	var responseCredentials *AwsCredentials
	var responseCode = 200
	if c, ok := responseBody.(ConsolemeCredentialResponseType); ok {
		responseCredentials = c.Credentials
	}
	if e, ok := responseBody.(ConsolemeCredentialErrorMessageType); ok {
		code, err := strconv.Atoi(e.Code)
		if err == nil {
			responseCode = code
		}
	}
	resp, err := json.Marshal(responseBody)
	if err != nil {
		return nil, err
	}
	var client HTTPClient
	client = &ClientMock{
		DoFunc: func(*http.Request) (*http.Response, error) {
			r := ioutil.NopCloser(bytes.NewReader(resp))
			return &http.Response{
				StatusCode: responseCode,
				Body:       r,
			}, nil
		},
		GetRoleCredentialsFunc: func(role string, ipRestrict bool) (*AwsCredentials, error) {
			if responseCredentials != nil {
				return responseCredentials, nil
			}
			return &AwsCredentials{RoleArn: role}, nil
		},
	}
	return client, nil
}
