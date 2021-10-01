package server

import (
	"encoding/json"
	"net/http"

	"github.com/netflix/weep/pkg/logging"
)

type httpError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

func errorResponse(w http.ResponseWriter, message, code string) {
	resp := httpError{
		Message: message,
		Code:    code,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logging.Log.Errorf("failed to write error response: %v", err)
	}
}
