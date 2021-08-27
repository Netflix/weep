package server

import (
	"encoding/json"
	"net/http"
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
		log.Errorf("failed to write error response: %v", err)
	}
}
