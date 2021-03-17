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

package metadata

import (
	"os"
	"os/user"
	"time"

	"github.com/netflix/weep/logging"
)

var (
	certCreationTime time.Time
	certFingerprint  string
	weepMethod       string
	weepStartupTime  time.Time
	log              = logging.GetLogger()
)

func init() {
	weepStartupTime = time.Now()
}

// GetInstanceInfo populates and returns an InstanceInfo, most likely to be used as
// request metadata.
func GetInstanceInfo() *InstanceInfo {
	currentTime := time.Now()
	currentInstanceInfo := &InstanceInfo{
		Hostname:              hostname(),
		Username:              username(),
		CertAgeSeconds:        elapsedSeconds(certCreationTime, currentTime),
		CertFingerprintSHA256: certFingerprint,
		WeepVersion:           Version,
		WeepMethod:            weepMethod,
	}
	return currentInstanceInfo
}

func StartupTime() string {
	return weepStartupTime.UTC().Format("2006-01-02T15:04:05Z")
}

func elapsedSeconds(startTime, endTime time.Time) int {
	if startTime.IsZero() || endTime.IsZero() {
		return 0
	}
	diff := endTime.Sub(startTime).Seconds()
	return int(diff)
}

// SetCertInfo stores the creation time and fingerprint of the in-use mTLS certificate.
func SetCertInfo(creationTime time.Time, fingerprint string) {
	certCreationTime = creationTime
	certFingerprint = fingerprint
}

func SetWeepMethod(command string) {
	weepMethod = command
}

func hostname() string {
	h, err := os.Hostname()
	if err != nil {
		log.Errorf("failed to get hostname: %v", err)
	}
	return h
}

func username() string {
	u, err := user.Current()
	if err != nil {
		log.Errorf("failed to get username: %v", err)
	}
	return u.Username
}
