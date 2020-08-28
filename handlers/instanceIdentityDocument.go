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

var (
	accountID string
)

func InstanceIdentityDocumentHandler(w http.ResponseWriter, r *http.Request) {

	awsArn, err := util.ArnParse(metadata.Role)

	if err != nil {
		accountID = "123456789012"
	} else {
		accountID = awsArn.AccountId
	}

	identityDocument := metadata.MetaDataInstanceIdentityDocumentResponse{
		DevpayProductCodes:      []string{},
		MarkerplaceProductCodes: []string{},
		PrivateIP:               "100.1.2.3",
		Version:                 "2017-09-30",
		InstanceID:              "i-12345",
		BillingProductCodes:     []string{},
		InstanceType:            "m5.large",
		AvailabilityZone:        "us-east-1a",
		KernelID:                "aki-fc8f11cc",
		RamdiskID:               "",
		AccountID:               accountID,
		Architecture:            "x86_64",
		ImageID:                 "ami-12345",
		PendingTime:             metadata.LastRenewal.UTC(), //.Format("2006-01-02T15:04:05Z"),
		Region:                  metadata.MetadataRegion,
	}

	b, err := json.Marshal(identityDocument)
	if err != nil {
		log.Error(err)
	}

	var out bytes.Buffer
	json.Indent(&out, b, "", "  ")
	fmt.Fprintln(w, out.String())
}
