package version

import (
	"bytes"
	"fmt"
	"github.com/netflix/weep/mtls"
)

var (
	GitCommit         string
	GitDescribe       string
	BuildDate         string
	Version           string
	VersionPrerelease string
)

// VersionInfo contains information about the program's version.
type VersionInfo struct {
	Revision          string
	Version           string
	VersionPrerelease string
	BuildDate         string
}

// GetVersion returns the program's version information via a VersionInfo pointer.
func GetVersion() *VersionInfo {
	ver := Version
	rel := VersionPrerelease

	if GitDescribe != "" {
		ver = GitDescribe
	}
	if GitDescribe == "" && rel == "" && VersionPrerelease != "" {
		rel = "dev"
	}

	return &VersionInfo{
		Revision:          GitCommit,
		Version:           ver,
		VersionPrerelease: rel,
		BuildDate:         BuildDate,
	}
}

func (c *VersionInfo) String() string {
	var versionString bytes.Buffer

	if Version == "" && VersionPrerelease == "" {
		fmt.Fprintf(&versionString, "weep (version unknown)")
	}
	fmt.Fprintf(&versionString, "weep v%s", c.Version)

	if c.VersionPrerelease != "" {
		fmt.Fprintf(&versionString, "-%s", c.VersionPrerelease)

	}

	if c.Revision != "" {
		fmt.Fprintf(&versionString, " (%s)", c.Revision)
	}

	fmt.Fprintf(&versionString, " Built on: %s", BuildDate)

	if mtls.EmbeddedConfigFile != "" {
		fmt.Fprintf(&versionString, " with embedded mTLS config %s", mtls.EmbeddedConfigFile)
	}

	return versionString.String()
}
