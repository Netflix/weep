package server

import (
	"encoding/json"
	"net/http"

	"github.com/netflix/weep/health"
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
	resp := healthcheckResponse{
		Status:  status,
		Message: reason,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Errorf("error writing healthcheck response: %v", err)
	}
}
