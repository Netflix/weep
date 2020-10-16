package main

import (
	"os"

	"github.com/netflix/weep/cmd"
	log "github.com/sirupsen/logrus"
)

func init() {
	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)
	log.SetReportCaller(true)
}

func main() {
	cmd.Execute()
}
