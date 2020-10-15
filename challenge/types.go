package challenge

type ConsolemeChallenge struct {
	ChallengeURL string `json:"challenge_url"`
	PollingUrl   string `json:"polling_url"`
}

type ConsolemeChallengeResponse struct {
	Status       string `json:"status"`
	EncodedJwt   string `json:"encoded_jwt"`
	CookieName   string `json:"cookie_name"`
	WantSecure   bool   `json:"secure"`
	WantHttpOnly bool   `json:"http_only"`
	SameSite     int    `json:"same_site"`
	Expires      int64  `json:"expiration"`
}
