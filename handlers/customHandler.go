package handlers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/netflix/weep/config"
)

func CustomHandler(w http.ResponseWriter, r *http.Request) {

	path := mux.Vars(r)["path"]

	for i := range config.Config.MetaData.Routes {
		if config.Config.MetaData.Routes[i].Path == path {
			fmt.Fprintln(w, config.Config.MetaData.Routes[i].Data)
		}
	}
}
