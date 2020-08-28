package handlers

import (
	"fmt"
	"net/http"
)

func BaseHandler(w http.ResponseWriter, r *http.Request) {

	baseMetadata := `ami-id
ami-launch-index
ami-manifest-path
block-device-mapping/
hostname
iam/
instance-action
instance-id
instance-type
kernel-id
local-hostname
local-ipv4
mac
metrics/
network/
placement/
profile
public-keys/
reservation-id
security-groups
services/`

	fmt.Fprint(w, baseMetadata)
}

func BaseVersionHandler(w http.ResponseWriter, r *http.Request) {

	baseVersionPath := `dynamic
meta-data
user-data`

	fmt.Fprintln(w, baseVersionPath)
}
