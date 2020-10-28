package consoleme

type AwsCredentials struct {
	AccessKeyId     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
	Expiration      int64  `json:"Expiration"`
	RoleArn         string `json:"RoleArn"`
}

type CredentialProcess struct {
	Version         int    `json:"Version"`
	AccessKeyId     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
	Expiration      string `json:"Expiration"`
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
