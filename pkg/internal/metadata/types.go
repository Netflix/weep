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

type InstanceInfo struct {
	Hostname              string `json:"hostname" yaml:"hostname"`
	Username              string `json:"username" yaml:"username"`
	CertAgeSeconds        int    `json:"cert_age_seconds,omitempty" yaml:"cert_age_seconds"`
	CertFingerprintSHA256 string `json:"cert_fingerprint_sha256,omitempty" yaml:"cert_fingerprint_sha256"`
	WeepVersion           string `json:"weep_version" yaml:"weep_version"`
	WeepMethod            string `json:"weep_method" yaml:"weep_method"`
}
