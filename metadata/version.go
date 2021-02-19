package metadata

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
