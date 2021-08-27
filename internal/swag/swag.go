package swag

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/netflix/weep/internal/httpAuth/mtls"
	"github.com/spf13/viper"
)

type SwagResponse struct {
	Name string `json:"name"`
}

func getClient() (*http.Client, error) {
	if viper.GetBool("swag.use_mtls") {
		return mtls.NewHTTPClient()
	}
	return http.DefaultClient, nil
}

func AccountName(id string) (string, error) {
	client, err := getClient()
	if err != nil {
		return "", err
	}
	urlStr := fmt.Sprintf("%s/1/accounts/%s", viper.GetString("swag.url"), id)
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		dec := json.NewDecoder(resp.Body)
		var r SwagResponse
		err := dec.Decode(&r)
		if err != nil {
			return "", err
		}
		return r.Name, nil
	}

	return "", nil
}
