package consoleme

type AwsCredentials struct {
	AccessKeyId     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
	Expiration      int64  `json:"Expiration"`
}

type ConsolemeCredentialResponseType struct {
	Credentials AwsCredentials `json:"Credentials"`
}

type ConsolemeCredentialRequestType struct {
	RequestedRole   string `json:"requested_role"`
	NoIpRestriciton bool   `json:"no_ip_restrictions"`
}

type ConsolemeCredentialErrorMessageType struct {
	Code          string `json:"code"`
	Message       string `json:"message"`
	RequestedRole string `json:"requested_role"`
	Exception     string `json:"exception"`
	RequestID     string `json:"request_id"`
}


type ConsolemeChallenge struct {
	ChallengeURL string `json:"challenge_url"`
	PollingUrl string `json:"polling_url"`
}

type ConsolemeChallengeResponse struct {
	Status string `json:"status"`
	EncodedJwt string `json:"encoded_jwt"`
	CookieName string `json:"cookie_name"`
	WantSecure bool `json:"secure"`
	WantHttpOnly bool `json:"http_only"`
	SameSite int `json:"same_site"`
	Expires int64 `json:"expiration"`
}
