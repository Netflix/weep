package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/netflix/weep/consoleme"
	"github.com/netflix/weep/metadata"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var credentialMap = make(map[string]consoleme.AwsCredentials)

func ECSMetadataServiceCredentialsHandler(w http.ResponseWriter, r *http.Request) {
	var client, err = consoleme.GetClient()
	if err != nil {
		log.Error(err)
		return
	}
	vars := mux.Vars(r)
	requestedRole := vars["role"]
	var Credentials consoleme.AwsCredentials

	val, ok := credentialMap[requestedRole]
	if ok {
		Credentials = val

		// Refresh credentials on demand if expired or within 10 minutes of expiry
		currentTime := time.Now()
		tm := time.Unix(Credentials.Expiration, 0)
		timeToRenew := tm.Add(-10 * time.Minute)
		if currentTime.After(timeToRenew) {
			Credentials, err = client.GetRoleCredentials(requestedRole, false)
			if err != nil {
				log.Error(err)
				return
			}
		}
	} else {
		Credentials, err = client.GetRoleCredentials(requestedRole, false)
		if err != nil {
			log.Error(err)
			return
		}
		credentialMap[requestedRole] = Credentials
	}

	tm := time.Unix(Credentials.Expiration, 0)

	credentials := metadata.ECSMetaDataCredentialResponse{
		AccessKeyId:     fmt.Sprintf("%s", Credentials.AccessKeyId),
		Expiration:      tm.UTC().Format("2006-01-02T15:04:05Z"),
		RoleArn:         Credentials.RoleArn,
		SecretAccessKey: fmt.Sprintf("%s", Credentials.SecretAccessKey),
		Token:           fmt.Sprintf("%s", Credentials.SessionToken),
	}

	b, err := json.Marshal(credentials)
	if err != nil {
		log.Error(err)
	}
	var out bytes.Buffer
	json.Indent(&out, b, "", "  ")
	fmt.Fprintln(w, out.String())
}
