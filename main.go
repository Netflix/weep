package main

import (
	"github.com/netflix/weep/cmd"
	log "github.com/sirupsen/logrus"
	"os"
)

func init() {
	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)
	log.SetReportCaller(true)
}

func main() {
	cmd.Execute()
}
