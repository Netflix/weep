package handlers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/netflix/weep/config"
)

func CustomHandler(w http.ResponseWriter, r *http.Request) {

	path := mux.Vars(r)["path"]

	for _, configRoute := range config.Config.MetaData.Routes {
		if configRoute.Path == path {
			fmt.Fprintln(w, configRoute.Path)
		}
	}
}
