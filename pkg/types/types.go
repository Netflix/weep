package types

import (
	"strconv"
	"sync"
	"time"
)

type Time time.Time

// MarshalJSON is used to convert the timestamp to JSON
func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(Time(t).Unix(), 10)), nil
}

// UnmarshalJSON is used to convert the timestamp from JSON
func (t *Time) UnmarshalJSON(s []byte) (err error) {
	r := string(s)
	q, err := strconv.ParseInt(r, 10, 64)
	if err != nil {
		return err
	}
	*(*time.Time)(t) = time.Unix(q, 0)
	return nil
}

// Add returns t with the provided duration added to it.
func (t Time) Add(d time.Duration) time.Time {
	return time.Time(t).Add(d)
}

// Unix returns t as a Unix time, the number of seconds elapsed
// since January 1, 1970 UTC. The result does not depend on the
// location associated with t.
func (t Time) Unix() int64 {
	return time.Time(t).Unix()
}

func (t Time) UTC() time.Time {
	return time.Time(t).UTC()
}

// Format returns t as a timestamp string with the provided layout.
func (t Time) Format(layout string) string {
	return time.Time(t).Format(layout)
}

// Time returns the JSON time as a time.Time instance in UTC
func (t Time) Time() time.Time {
	return time.Time(t).UTC()
}

// String returns t as a formatted string
func (t Time) String() string {
	return t.Time().String()
}

type CredentialProcess struct {
	Version         int    `json:"Version"`
	AccessKeyId     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
	Expiration      string `json:"Expiration"`
}

type Credentials struct {
	Role                string
	NoIpRestrict        bool
	metaDataCredentials *Credentials
	MetadataRegion      string
	LastRenewal         Time
	mu                  sync.Mutex
}

// RoleDetails represents the response structure of Weep's model for detailed eligible roles
type RoleDetails struct {
	Arn           string `json:"arn"`
	AccountNumber string `json:"account_id"`
	AccountName   string `json:"account_friendly_name"`
	RoleName      string `json:"role_name"`
	Apps          struct {
		AppDetails []AppDetails `json:"app_details"`
	} `json:"apps"`
}

// AppDetails represents the structure of details returned by ConsoleMe about a single app
type AppDetails struct {
	Name     string `json:"name"`
	Owner    string `json:"owner"`
	OwnerURL string `json:"owner_url"`
	AppURL   string `json:"app_url"`
}
