package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/netflix/weep/metadata"
	"github.com/netflix/weep/util"
)

func IamInfoHandler(w http.ResponseWriter, r *http.Request) {

	awsArn, _ := util.ArnParse(metadata.Role)

	awsArn.ResourceType = "instance-profile"

	iamInfo := metadata.MetaDataIamInfoResponse{
		Code:               "Success",
		LastUpdated:        metadata.LastRenewal.UTC().Format("2006-01-02T15:04:05Z"),
		InstanceProfileARN: awsArn.ArnString(),
		InstanceProfileID:  "AIPAI",
	}

	b, err := json.Marshal(iamInfo)
	if err != nil {
		log.Error(err)
	}
	var out bytes.Buffer
	json.Indent(&out, b, "", "  ")
	fmt.Fprintln(w, out.String())
}
