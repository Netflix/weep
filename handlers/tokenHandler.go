package handlers

import (
	"fmt"
	"net/http"
)

const staticToken = "AQAEANQlVdnIoNfmJQHofbSTjkIm8eoMIBZZZX05Xk9jLiFuJuL2_A=="

func TokenHandler(w http.ResponseWriter, r *http.Request) {
	// Returning a static token allows us to support IMDSv2 with minimal effort.
	fmt.Fprint(w, staticToken)
}
