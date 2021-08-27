package aws

import "github.com/netflix/weep/internal/types"

type Credentials struct {
	AccessKeyId     string     `json:"AccessKeyId"`
	SecretAccessKey string     `json:"SecretAccessKey"`
	SessionToken    string     `json:"SessionToken"`
	Expiration      types.Time `json:"Expiration"`
	RoleArn         string     `json:"RoleArn"`
}
