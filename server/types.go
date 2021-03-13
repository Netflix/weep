package server

type MetaDataCredentialResponse struct {
	Code            string
	LastUpdated     string
	Type            string
	AccessKeyId     string
	SecretAccessKey string
	Token           string
	Expiration      string
}

type ECSMetaDataCredentialResponse struct {
	AccessKeyId     string
	SecretAccessKey string
	Token           string
	Expiration      string
	RoleArn         string
}

type MetaDataIamInfoResponse struct {
	Code               string `json:"Code"`
	LastUpdated        string `json:"LastUpdated"`
	InstanceProfileARN string `json:"InstanceProfileArn"`
	InstanceProfileID  string `json:"InstanceProfileId"`
}

type MetaDataInstanceIdentityDocumentResponse struct {
	DevpayProductCodes      []string `json:"devpayProductCodes"`
	MarkerplaceProductCodes []string `json:"marketplaceProductCodes"`
	PrivateIP               string   `json:"privateIp"`
	Version                 string   `json:"version"`
	InstanceID              string   `json:"instanceId"`
	BillingProductCodes     []string `json:"billingProducts"`
	InstanceType            string   `json:"instanceType"`
	AvailabilityZone        string   `json:"availabilityZone"`
	KernelID                string   `json:"kernelId"`
	RamdiskID               string   `json:"ramdiskId"`
	AccountID               string   `json:"accountId"`
	Architecture            string   `json:"architecture"`
	ImageID                 string   `json:"imageId"`
	PendingTime             string   `json:"pendingTime"`
	Region                  string   `json:"region"`
}
