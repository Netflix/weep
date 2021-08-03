package session

import (
	"github.com/netflix/weep/logging"
)

var sessions *tokenCache
var log = logging.GetLogger()

func init() {
	sessions = createCache()
}

func GenerateToken(role string, ttlSeconds int) string {
	return sessions.generateToken(role, ttlSeconds)
}

func CheckToken(token string) (bool, int) {
	return sessions.checkToken(token)
}
