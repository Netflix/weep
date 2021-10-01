package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/netflix/weep/pkg/logging"
	"github.com/netflix/weep/pkg/reachability"

	"github.com/netflix/weep/pkg/health"
)

type healthcheckResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	healthy, reason := health.WeepStatus.Get()
	var status int
	if healthy {
		status = http.StatusOK
	} else {
		status = http.StatusInternalServerError
	}

	reachabilityFlag := r.URL.Query().Get("reachability")
	if b, err := strconv.ParseBool(reachabilityFlag); err == nil && b {
		reachability.Notify()
	}

	resp := healthcheckResponse{
		Status:  status,
		Message: reason,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		logging.Log.Errorf("error writing healthcheck response: %v", err)
	}
}
