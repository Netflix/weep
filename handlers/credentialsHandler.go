package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/netflix/weep/metadata"
	"github.com/netflix/weep/util"
)

func RoleHandler(w http.ResponseWriter, r *http.Request) {

	arn, _ := util.ArnParse(metadata.Role)
	if arn != nil {
		fmt.Fprint(w, arn.Resource)
	}
}

func CredentialsHandler(w http.ResponseWriter, r *http.Request) {

	tm := time.Unix(metadata.MetaDataCredentials.Expiration, 0)

	credentials := metadata.MetaDataCredentialResponse{
		Code:            "Success",
		LastUpdated:     metadata.LastRenewal.UTC().Format("2006-01-02T15:04:05Z"),
		Type:            "AWS-HMAC",
		AccessKeyId:     fmt.Sprintf("%s", metadata.MetaDataCredentials.AccessKeyId),
		SecretAccessKey: fmt.Sprintf("%s", metadata.MetaDataCredentials.SecretAccessKey),
		Token:           fmt.Sprintf("%s", metadata.MetaDataCredentials.SessionToken),
		Expiration:      tm.UTC().Format("2006-01-02T15:04:05Z"),
	}

	b, err := json.Marshal(credentials)
	if err != nil {
		log.Error(err)
	}
	var out bytes.Buffer
	json.Indent(&out, b, "", "  ")
	fmt.Fprintln(w, out.String())
}
