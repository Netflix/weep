package reachability

import (
	"time"
)

const maxWaitTimeSeconds = 3

var c chan struct{}

func init() {
	c = make(chan struct{})
}

// Notify will signal the reachability package that some reachability test
// were received by this Weep instance
func Notify() {
	// Only sends when there is already some receiver waiting. Never blocks.
	select {
	case c <- struct{}{}:
	default:
	}
}

func wait() bool {
	timeout := make(chan struct{})
	go func() {
		time.Sleep(maxWaitTimeSeconds * time.Second)
		timeout <- struct{}{}
	}()

	select {
	case <-c:
		// Received a rechability test
		return true
	case <-timeout:
		// Timed out, move on
	}

	return false
}
