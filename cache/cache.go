package cache

import (
	"strings"

	"github.com/netflix/weep/creds"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var GlobalCache WeepCache

func init() {
	switch cacheType := viper.GetString("cache.type"); cacheType {
	case "memory":
		GlobalCache = NewMemoryCache()
	case "file":
		// TODO
	default:
		log.Fatal("invalid cache type specified")
	}
}

type WeepCache interface {
	Get(role string, assumeChain []string) (*creds.RefreshableProvider, error)
	GetOrSet(client *creds.Client, role string, region string, assumeChain []string) (*creds.RefreshableProvider, error)
	SetDefault(client *creds.Client, role string, region string, assumeChain []string) error
	GetDefault() (*creds.RefreshableProvider, error)
}

// getSlug returns a string unique to a particular combination of a role and chain of roles to assume.
func getSlug(role string, assume []string) string {
	elements := append([]string{role}, assume...)
	return strings.Join(elements, "/")
}
