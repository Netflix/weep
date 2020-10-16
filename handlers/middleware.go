package handlers

import (
	"math/rand"
	"net/http"
	"strconv"

	"github.com/netflix/weep/metadata"
	log "github.com/sirupsen/logrus"
)

func MetaDataServiceMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("ETag", strconv.FormatInt(rand.Int63n(10000000000), 10))
		w.Header().Set("Last-Modified", metadata.LastRenewal.UTC().Format("2006-01-02T15:04:05Z"))
		w.Header().Set("Server", "EC2ws")
		w.Header().Set("Content-Type", "text/plain")

		ua := string(r.Header.Get("User-Agent"))
		metadataVersion := 1
		token_ttl := r.Header.Get("X-Aws-Ec2-Metadata-Token-Ttl-Seconds")
		token := r.Header.Get("X-aws-ec2-metadata-token")
		// If either of these request headers exist, we can be reasonably confident that the request is for IMDSv2.
		// `X-Aws-Ec2-Metadata-Token-Ttl-Seconds` is used when requesting a token
		// `X-aws-ec2-metadata-token` is used to pass the token to the metadata service
		// Weep uses a static token, and does not perform any token validation.
		if token != "" || token_ttl != "" {
			metadataVersion = 2
		}

		if !checkUserAgent(ua) {
			log.WithFields(log.Fields{
				"user-agent":       ua,
				"path":             r.URL.Path,
				"metadata_version": metadataVersion,
			}).Info("You are using a SDK that does not support User-Agents that Netflix wants")
		} else {
			log.WithFields(log.Fields{
				"user-agent":       ua,
				"path":             r.URL.Path,
				"metadata_version": metadataVersion,
			}).Info()
		}
		next.ServeHTTP(w, r)
	}
}
