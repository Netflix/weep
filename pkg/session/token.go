package session

var sessions *tokenCache

func init() {
	sessions = createCache()
}

func GenerateToken(role string, ttlSeconds int) string {
	return sessions.generateToken(role, ttlSeconds)
}

func CheckToken(token string) (bool, int) {
	return sessions.checkToken(token)
}
