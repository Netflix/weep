package consoleme

import (
	"encoding/json"

	"github.com/netflix/weep/pkg/aws"
	"github.com/netflix/weep/pkg/metadata"
)

type errorResponse struct {
	Code          string `json:"code"`
	Message       string `json:"message"`
	RequestedRole string `json:"requested_role"`
	Exception     string `json:"exception"`
	RequestID     string `json:"request_id"`
}

// webResponse represents the response structure of ConsoleMe's model WebResponse
type webResponse struct {
	Status      string                     `json:"status"`
	Reason      string                     `json:"reason"`
	RedirectURL string                     `json:"redirect_url"`
	StatusCode  int                        `json:"status_code"`
	Message     string                     `json:"message"`
	Errors      []string                   `json:"errors"`
	Data        map[string]json.RawMessage `json:"data"`
}

type credentialResponse struct {
	Credentials *aws.Credentials `json:"Credentials"`
}

type credentialRequest struct {
	RequestedRole  string                 `json:"requested_role"`
	NoIpRestricton bool                   `json:"no_ip_restrictions"`
	Metadata       *metadata.InstanceInfo `json:"metadata,omitempty"`
}
