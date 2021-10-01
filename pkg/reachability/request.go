package reachability

import (
	"net/http"
	"os"

	"github.com/netflix/weep/pkg/logging"
)

// TestReachability sends a GET request to the address IMDS is expected to run, while checks
// whether this test was received by this same Weep instance, otherwise logs a warning
func TestReachability() {
	go func() {
		logging.Log.Debug("Doing a healthcheck request on 169.254.169.254")
		resp, err := http.Get("http://169.254.169.254/healthcheck?reachability=1")

		// A response can be successful but have being served by another process on the
		// IMDS port/an actual IMDS. So we prefer relying on the reachability signal (which
		// means this same process received a reachability test).

		if err != nil {
			logging.Log.WithField("err", err).Debug("Received an error from healthcheck route")
		} else {
			logging.Log.WithField("status", resp.StatusCode).Debug("Received a response from healthcheck route")
		}
	}()

	received := wait()
	if received {
		logging.Log.Info("Reachability test was successful")
	} else {
		logging.Log.Warningf(
			"Reachability test was unsuccessful. Looks like we aren't being served in 169.254.169.254. Did you `%s setup`?",
			os.Args[0],
		)
	}
}
