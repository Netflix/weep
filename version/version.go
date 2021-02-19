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

package version

import (
	"bytes"
	"fmt"

	"github.com/netflix/weep/config"
)

var (
	Commit  string
	Date    string
	Version string
)

// VersionInfo contains information about the program's version.
type VersionInfo struct {
	Revision  string
	Version   string
	BuildDate string
}

// GetVersion returns the program's version information via a VersionInfo pointer.
func GetVersion() *VersionInfo {
	ver := Version

	return &VersionInfo{
		Revision:  Commit,
		Version:   ver,
		BuildDate: Date,
	}
}

func (c *VersionInfo) String() string {
	var versionString bytes.Buffer

	if Version == "" {
		_, _ = fmt.Fprintf(&versionString, "weep (version unknown)")
	}
	_, _ = fmt.Fprintf(&versionString, "weep %s", c.Version)

	if c.Revision != "" {
		_, _ = fmt.Fprintf(&versionString, " (%s)", c.Revision)
	}

	_, _ = fmt.Fprintf(&versionString, " Built on: %s", Date)

	if config.EmbeddedConfigFile != "" {
		_, _ = fmt.Fprintf(&versionString, " with embedded config %s", config.EmbeddedConfigFile)
	}

	return versionString.String()
}
