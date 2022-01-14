package httpAuth

import (
	"fmt"
	"net/http"

	"github.com/netflix/weep/pkg/httpAuth/challenge"
	"github.com/netflix/weep/pkg/httpAuth/custom"
	"github.com/netflix/weep/pkg/httpAuth/mtls"
	"github.com/spf13/viper"
)

func GetAuthenticatedClient() (*http.Client, error) {
	authenticationMethod := viper.GetString("authentication_method")
	consoleMeUrl := viper.GetString("consoleme_url")
	if custom.UseCustom() {
		return custom.NewHTTPClient()
	} else if authenticationMethod == "mtls" {
		return mtls.NewHTTPClient()
	} else if authenticationMethod == "challenge" {
		err := challenge.RefreshChallenge()
		if err != nil {
			return nil, err
		}
		return challenge.NewHTTPClient(consoleMeUrl)
	}
	return nil, fmt.Errorf("Authentication method unsupported or not provided.")
}
