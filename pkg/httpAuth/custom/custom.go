package custom

import "net/http"

var isOverridden bool
var clientFactoryOverride ClientFactory
var preflightFunctions = make([]RequestPreflight, 0)

type ClientFactory func() (*http.Client, error)

func UseCustom() bool {
	return isOverridden
}

func NewHTTPClient() (*http.Client, error) {
	return clientFactoryOverride()
}

// RegisterClientFactory overrides Weep's standard config-based ConsoleMe client
// creation with a ClientFactory. This function will be called during the creation
// of all ConsoleMe clients.
func RegisterClientFactory(factory ClientFactory) {
	clientFactoryOverride = factory
	isOverridden = true
}

type RequestPreflight func(req *http.Request) error

// RegisterRequestPreflight adds a RequestPreflight function which will be called in the
// order of registration during the creation of a ConsoleMe request.
func RegisterRequestPreflight(preflight RequestPreflight) {
	preflightFunctions = append(preflightFunctions, preflight)
}

func RunPreflightFunctions(req *http.Request) error {
	var err error
	if preflightFunctions != nil {
		for _, preflight := range preflightFunctions {
			if err = preflight(req); err != nil {
				return err
			}
		}
	}
	return nil
}
