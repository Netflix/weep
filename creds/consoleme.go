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
	"syscall"
	"time"

	"github.com/netflix/weep/logging"

	werrors "github.com/netflix/weep/errors"
	"github.com/spf13/viper"

	"github.com/netflix/weep/httpAuth/challenge"
	"github.com/netflix/weep/httpAuth/mtls"

	"github.com/pkg/errors"

	"github.com/netflix/weep/version"
)

var log = logging.GetLogger()
var clientVersion = fmt.Sprintf("%s", version.Version)

var userAgent = "weep/" + clientVersion + " Go-http-client/1.1"

type Account struct {
}

// HTTPClient is the interface we expect HTTP clients to implement.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// Client represents a ConsoleMe client.
type Client struct {
	Httpc HTTPClient
	Host  string
}

// GetClient creates an authenticated ConsoleMe client
func GetClient() (*Client, error) {
	var client *Client
	consoleMeUrl := viper.GetString("consoleme_url")
	authenticationMethod := viper.GetString("authentication_method")

	if authenticationMethod == "mtls" {
		mtlsClient, err := mtls.NewHTTPClient()
		if err != nil {
			return client, err
		}
		client, err = NewClient(consoleMeUrl, mtlsClient)
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
		client, err = NewClient(consoleMeUrl, httpClient)
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
func NewClient(hostname string, httpc HTTPClient) (*Client, error) {
	if len(hostname) == 0 {
		return nil, errors.New("hostname cannot be empty string")
	}

	if httpc == nil {
		httpc = &http.Client{Transport: defaultTransport()}
	}

	c := &Client{
		Httpc: httpc,
		Host:  hostname,
	}

	return c, nil
}

func (c *Client) buildRequest(method string, resource string, body io.Reader) (*http.Request, error) {
	urlStr := c.Host + "/api/v1" + resource

	return http.NewRequest(method, urlStr, body)
}

// do invokes an HTTP request, and returns the response. This also sets the
// User-Agent of the client.
func (c *Client) do(req *http.Request) (*http.Response, error) {
	req.Header.Set(("User-Agent"), userAgent)
	req.Header.Add("Content-Type", "application/json")

	return c.Httpc.Do(req)
}

// accounts returns all accounts, and allows you to filter the accounts by sub-resources
// like: /accounts/service/support
func (c *Client) Roles() ([]string, error) {
	var cmCredentialErrorMessageType ConsolemeCredentialErrorMessageType
	req, err := c.buildRequest(http.MethodGet, "/get_roles?all=true", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build request")
	}

	// Add URL Parameters
	q := url.Values{}
	q.Add("all", "true")
	req.URL.RawQuery = q.Encode()

	resp, err := c.do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to action request")
	}

	defer resp.Body.Close()
	document, err := ioutil.ReadAll(resp.Body)
	// Handle invalid_jwt error from ConsoleMe by deleting local credentials and asking user to retry
	if resp.StatusCode == 403 {
		if err := json.Unmarshal(document, &cmCredentialErrorMessageType); err != nil {
			return nil, err
		}
		if cmCredentialErrorMessageType.Code == "invalid_jwt" {
			log.Errorf("Authentication is invalid or has expired. Please restart weep to re-authenticate.")
			err = challenge.DeleteLocalWeepCredentials()
			if err != nil {
				return nil, errors.Wrap(err, "Failed to delete invalid Weep jwt")
			}
		}
	}
	// Handle non-200 errors
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected HTTP status %s, want 200. Body: %s", resp.Status, string(document))
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	var roles []string
	if err := json.Unmarshal(document, &roles); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal JSON")
	}

	return roles, nil
}

func (c *Client) GetRoleCredentials(role string, ipRestrict bool) (*AwsCredentials, error) {
	var credentialsResponse ConsolemeCredentialResponseType
	var cmCredentialErrorMessageType ConsolemeCredentialErrorMessageType

	cmCredRequest := ConsolemeCredentialRequestType{
		RequestedRole:   role,
		NoIpRestriciton: ipRestrict,
	}

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(cmCredRequest)
	if err != nil {
		return credentialsResponse.Credentials, errors.Wrap(err, "failed to create request body")
	}

	req, err := c.buildRequest(http.MethodPost, "/get_credentials", b)
	if err != nil {
		return credentialsResponse.Credentials, errors.Wrap(err, "failed to build request")
	}

	resp, err := c.do(req)
	if err != nil {
		return credentialsResponse.Credentials, errors.Wrap(err, "failed to action request")
	}

	defer resp.Body.Close()
	document, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		if resp.StatusCode == 403 {
			if err != nil {
				return credentialsResponse.Credentials, errors.Wrap(err, "failed to read response body")
			}
			if err := json.Unmarshal(document, &cmCredentialErrorMessageType); err != nil {
				return credentialsResponse.Credentials, errors.Wrap(err, "failed to unmarshal JSON")
			}
			if cmCredentialErrorMessageType.Code == "905" {
				return credentialsResponse.Credentials, fmt.Errorf(viper.GetString("mtls_settings.old_cert_message"))
			}
			if cmCredentialErrorMessageType.Code == "invalid_jwt" {
				log.Errorf("Authentication is invalid or has expired. Please restart weep to re-authenticate.")
				err = challenge.DeleteLocalWeepCredentials()
				if err != nil {
					fmt.Println(err.Error())
				}
				syscall.Exit(1)
			}
		}
		return credentialsResponse.Credentials, fmt.Errorf("unexpected HTTP status %s, want 200. Response: %s", resp.Status, string(document))
	}

	if err != nil {
		return credentialsResponse.Credentials, errors.Wrap(err, "failed to read response body")
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
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (c *ClientMock) Do(req *http.Request) (*http.Response, error) {
	return c.DoFunc(req)
}

func GetTestClient(responseBody interface{}) (*Client, error) {
	resp, err := json.Marshal(responseBody)
	if err != nil {
		return nil, err
	}
	client := &Client{
		Httpc: &ClientMock{
			DoFunc: func(*http.Request) (*http.Response, error) {
				r := ioutil.NopCloser(bytes.NewReader(resp))
				return &http.Response{
					StatusCode: 200,
					Body:       r,
				}, nil
			},
		},
	}
	return client, nil
}
