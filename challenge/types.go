/*
 * Copyright 2020 Netflix, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
	User         string `json:"user"`
}
