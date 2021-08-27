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

package server

import (
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/netflix/weep/pkg/session"
	"github.com/netflix/weep/pkg/util"

	"github.com/spf13/viper"

	"github.com/sirupsen/logrus"
)

// InstanceMetadataMiddleware is a convenience wrapper that chains TokenMiddleware, BrowserFilterMiddleware, and AWSHeaderMiddleware
func InstanceMetadataMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return TokenMiddleware(TaskMetadataMiddleware(next))
}

// TaskMetadataMiddleware is a convenience wrapper that chains BrowserFilterMiddleware and AWSHeaderMiddleware
func TaskMetadataMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return BrowserFilterMiddleware(AWSHeaderMiddleware(next))
}

func TokenMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var remainingTtl int
		var ok bool

		token := r.Header.Get("x-aws-ec2-metadata-token")
		if token != "" {
			if ok, remainingTtl = session.CheckToken(token); !ok {
				log.Debug("token invalid")
				util.WriteError(w, "invalid session token", http.StatusForbidden)
				return
			}
		} else if token == "" && viper.GetBool("server.enforce_imdsv2") {
			log.Info("request forbidden, imdsv2 required")
			util.WriteError(w, "IMDSv2 required, please upgrade your SDK or CLI", http.StatusForbidden)
			return
		}

		// Return the token's remaining TTL in a header
		if remainingTtl > 0 {
			w.Header().Set("X-Aws-Ec2-Metadata-Token-Ttl-Seconds", strconv.Itoa(remainingTtl))
		}
		next.ServeHTTP(w, r)
	}
}

func AWSHeaderMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("ETag", strconv.FormatInt(rand.Int63n(10000000000), 10))
		w.Header().Set("Last-Modified", time.Now().UTC().Format("2006-01-02T15:04:05Z")) // TODO: set this to cred refresh time
		w.Header().Set("Server", "EC2ws")
		w.Header().Set("Content-Type", "text/plain")

		ua := r.Header.Get("User-Agent")
		metadataVersion := 1
		tokenTtl := r.Header.Get("X-Aws-Ec2-Metadata-Token-Ttl-Seconds")
		token := r.Header.Get("X-aws-ec2-metadata-token")
		// If either of these request headers exist, we can be reasonably confident that the request is for IMDSv2.
		// `X-Aws-Ec2-Metadata-Token-Ttl-Seconds` is used when requesting a token
		// `X-aws-ec2-metadata-token` is used to pass the token to the metadata service
		if token != "" || tokenTtl != "" {
			metadataVersion = 2
		}

		log.WithFields(logrus.Fields{
			"user-agent":       ua,
			"path":             r.URL.Path,
			"metadata_version": metadataVersion,
		}).Info()
		next.ServeHTTP(w, r)
	}
}

// allowedHosts is a map used to look up Host headers for the purpose of rejecting requests
// for hosts that are not allowed
var allowedHosts = map[string]bool{
	"localhost":       true, // localhost
	"127.0.0.1":       true, // localhost
	"169.254.169.254": true, // IMDS IP
}

// deniedHeaders is a list of headers that will cause a 403 if present at all
var deniedHeaders = map[string]bool{
	"referrer":        true,
	"origin":          true,
	"x-forwarded-for": true,
}

// BrowserFilterMiddleware is a middleware designed mitigate risks related to DNS rebinding,
// cross site request forgery, and any other traffic from a well behaved modern web browser
func BrowserFilterMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check User-Agent
		// If User-Agent has Mozilla in it, this is almost certainly a browser request
		userAgent := r.Header.Get("User-Agent")
		userAgent = strings.ToLower(userAgent)
		if strings.Contains(userAgent, "mozilla") {
			log.Warn("bad user-agent detected")
			util.WriteError(w, "forbidden", http.StatusForbidden)
			return
		}

		// Check for presence of deniedHeaders
		// These also indicate a likely browser request
		for h, _ := range r.Header {
			if deniedHeaders[strings.ToLower(h)] {
				log.Warnf("%s header detected", h)
				util.WriteError(w, "forbidden", http.StatusForbidden)
				return
			}
		}

		// Check host header
		// This should only be 127.0.0.1 or 169.254.169.254
		if host := r.Header.Get("Host"); host != "" && !allowedHosts[strings.ToLower(host)] {
			log.Warn("bad host detected")
			util.WriteError(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}
}
