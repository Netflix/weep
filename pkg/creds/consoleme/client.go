package consoleme

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/netflix/weep/pkg/types"

	"github.com/netflix/weep/pkg/aws"
	"github.com/netflix/weep/pkg/config"
	"github.com/netflix/weep/pkg/creds/consoleme/challenge"
	werrors "github.com/netflix/weep/pkg/errors"
	"github.com/netflix/weep/pkg/logging"
	"github.com/netflix/weep/pkg/metadata"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	weepVersion = fmt.Sprintf("%s", metadata.Version)
	userAgent   = "weep/" + weepVersion + " Go-http-client/1.1"
)

func buildRequest(ctx context.Context, method, resource string, body io.Reader, apiPrefix string) (*http.Request, error) {
	baseURL := viper.GetString("consoleme_url")
	urlStr := baseURL + apiPrefix + resource
	req, err := http.NewRequestWithContext(ctx, method, urlStr, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	return req, nil
}

func parseError(statusCode int, rawErrorResponse []byte) error {
	var errorResponse errorResponse
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
		logging.Log.Errorf("Authentication is invalid or has expired. Please restart weep to re-authenticate.")
		err := challenge.DeleteLocalWeepCredentials()
		if err != nil {
			logging.Log.Errorf("failed to delete credentials: %v", err)
		}
		return werrors.InvalidJWT
	default:
		return fmt.Errorf("unexpected HTTP status %d, want 200. Response: %s", statusCode, string(rawErrorResponse))
	}
}

func parseErrorRaw(rawErrorResponse []byte) error {
	var errorResponse webResponse
	if err := json.Unmarshal(rawErrorResponse, &errorResponse); err != nil {
		return errors.Wrap(err, "failed to unmarshal JSON")
	}
	return fmt.Errorf(strings.Join(errorResponse.Errors, "\n"))
}

func retrieveCredentials(ctx context.Context, client *http.Client, searchString string, ipRestrict bool) (*aws.Credentials, error) {
	var credentialsResponse credentialResponse

	cmCredRequest := credentialRequest{
		RequestedRole:  searchString,
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

	req, err := buildRequest(ctx, http.MethodPost, "/get_credentials", b, "/api/v1")
	if err != nil {
		return credentialsResponse.Credentials, errors.Wrap(err, "failed to build request")
	}

	resp, err := client.Do(req)
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

func retrieveRoles(ctx context.Context, client *http.Client) ([]string, error) {
	req, err := buildRequest(ctx, http.MethodGet, "/get_roles", nil, "/api/v1")
	if err != nil {
		return nil, errors.Wrap(err, "failed to build request")
	}

	// Add URL Parameters
	q := url.Values{}
	q.Add("all", "true")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
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

func retrieveRolesExtended(ctx context.Context, client *http.Client) ([]types.RoleDetails, error) {
	req, err := buildRequest(ctx, http.MethodGet, "/get_roles", nil, "/api/v2")
	if err != nil {
		return nil, errors.Wrap(err, "failed to build request")
	}

	// Add URL Parameters
	q := url.Values{}
	q.Add("all", "true")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
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

	var responseParsed webResponse
	if err := json.Unmarshal(document, &responseParsed); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal JSON")
	}
	var roles []types.RoleDetails
	if err = json.Unmarshal(responseParsed.Data["roles"], &roles); err != nil {
		return nil, werrors.UnexpectedResponseType
	}

	return roles, nil
}

func retrieveResourceURL(ctx context.Context, client *http.Client, arn string) (string, error) {
	req, err := buildRequest(ctx, http.MethodGet, "/get_resource_url", nil, "/api/v2")
	if err != nil {
		return "", errors.Wrap(err, "failed to build request")
	}

	// Add URL Parameters
	q := url.Values{}
	q.Add("arn", arn)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to action request")
	}

	defer resp.Body.Close()
	document, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read response body")
	}
	if resp.StatusCode != http.StatusOK {
		return "", parseErrorRaw(document)
	}
	var responseParsed webResponse
	if err := json.Unmarshal(document, &responseParsed); err != nil {
		return "", errors.Wrap(err, "failed to unmarshal JSON")
	}
	var respURL string
	if err = json.Unmarshal(responseParsed.Data["url"], &respURL); err != nil {
		return "", werrors.UnexpectedResponseType
	}
	return config.BaseWebURL() + respURL, nil
}
