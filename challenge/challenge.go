package challenge

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"github.com/netflix/weep/config"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func NewHTTPClient(consolemeUrl string) (*http.Client, error) {
	if !HasValidJwt() {
		return nil, errors.New("Your authentication to ConsoleMe has expired. Please restart weep.")
	}
	var challenge ConsolemeChallengeResponse
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil { return nil, err }
	credentialsPath, err := getCredentialsPath()
	if err != nil { return nil, err }
	challengeBody, err := ioutil.ReadFile(credentialsPath)
	if err != nil { return nil, err }
	err = json.Unmarshal(challengeBody, &challenge)
	if err != nil { return nil, err }
	cookies := []*http.Cookie{{
		Name:     challenge.CookieName,
		Value:    challenge.EncodedJwt,
		Secure:   challenge.WantSecure,
		HttpOnly: challenge.WantHttpOnly,
		SameSite: http.SameSite(challenge.SameSite),
		Expires:  time.Unix(challenge.Expires, 0),
	},
	}
	consoleMeUrlParsed, err := url.Parse(consolemeUrl)
	if err != nil { return nil, err }
	jar.SetCookies(consoleMeUrlParsed, cookies)
	if err != nil { return nil, err }
	client := &http.Client{
		Jar: jar,
	}

	return client, err
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isWSL() bool {
	if fileExists("/proc/sys/kernel/osrelease") {
		if osrelease, err := ioutil.ReadFile("/proc/sys/kernel/osrelease"); err == nil {
			if strings.Contains(strings.ToLower(string(osrelease)), "microsoft") {
				return true
			}
		}
	}
	return false
}

func poll(pollingUrl string) (*ConsolemeChallengeResponse, error) {
	var pollResponse ConsolemeChallengeResponse
	var pollResponseBody []byte
	timeout := time.After(2 * time.Minute)
	tick := time.Tick(3 * time.Second)
	req, err := http.NewRequest("GET", pollingUrl, nil)
	if err != nil { return nil, err }
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	// Keep trying until we're timed out or got a result or got an error
	for {
		select {
		case <-timeout:
			return nil, errors.New("*** Unable to validate Challenge Response after 2 minutes. Quitting. ***")
		case <-tick:
			resp, err := client.Do(req)
			if err != nil { return nil, err }
			defer resp.Body.Close()
			if resp.Body != nil {
				pollResponseBody, err = ioutil.ReadAll(resp.Body)
				err := json.Unmarshal(pollResponseBody, &pollResponse)
				if err != nil { return nil, err }
				if pollResponse.Status == "success" {
					return &pollResponse, nil
				}
			}
		}
	}
}

func getCredentialsPath() (string, error) {
	currentUser, err := user.Current()
	if err != nil { return "", err }
	weepDir := filepath.Join(currentUser.HomeDir, ".weep")
	// Setup the directories where we will be writing credentials
	if _, err := os.Stat(weepDir); os.IsNotExist(err) {
		_ = os.Mkdir(weepDir, 0700)
	} else {
		_ = os.Chmod(weepDir, 0700)
	}
	credentialsPath := filepath.Join(weepDir, ".credentials")
	return credentialsPath, nil
}

func HasValidJwt() bool {
	var challenge ConsolemeChallengeResponse
	credentialPath, err := getCredentialsPath()
	if err != nil {
		return false
	}
	challengeBody, err := ioutil.ReadFile(credentialPath)
	if err != nil {
		return false
	}
	err = json.Unmarshal(challengeBody, &challenge)
	if err != nil {
		return false
	}
	now := time.Now()
	expires := time.Unix(challenge.Expires, 0)
	if now.After(expires) {
		return false
	}
	return true
}

func RefreshChallenge() error {
	// If credentials are still valid, no need to refresh them.
	if HasValidJwt() {
		return nil
	}
	// Step 1: Make unauthed request to ConsoleMe challenge endpoint and get a challenge challenge
	if config.Config.ChallengeSettings.User == "" {
		log.Fatalf(
			"Invalid configuration. You must define challenge_settings.user as the ",
			"user you wish to authenticate as.",
		)
	}
	var consoleMeChallengeGeneratorEndpoint string = fmt.Sprintf(
		"%s/noauth/v1/challenge_generator/%s",
		config.Config.ConsoleMeUrl,
		config.Config.ChallengeSettings.User,
	)
	var challenge ConsolemeChallenge
	req, err := http.NewRequest("GET", consoleMeChallengeGeneratorEndpoint, nil)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()
	tokenResponseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil { return err }

	if err := json.Unmarshal(tokenResponseBody, &challenge); err != nil {
		return err
	}

	log.Infof("Opening browser to Challenge URL location: %s", challenge.ChallengeURL)

	// Step 2: Make a web request to ChallengeUrl with user's browser
	var openUrlCommand []string = nil
	switch runtime.GOOS {
	case "darwin":
		openUrlCommand = []string{"open"}
	case "linux":
		if isWSL() {
			openUrlCommand = []string{"cmd.exe", "/C", "start"}
		} else {
			openUrlCommand = []string{"xdg-open"}
		}
	case "windows":
		openUrlCommand = []string{"cmd", "/C", "start"}
	}

	if openUrlCommand != nil {
		cmd := exec.Command(openUrlCommand[0], append(openUrlCommand[1:], challenge.ChallengeURL)...)
		err = cmd.Start()
		if err == nil {
			err = cmd.Wait()
		}
		if err != nil {
			log.Errorf("Failed to open browser with '%s': %s.",
				openUrlCommand[0], err.Error())
			log.Infoln("*** Could not launch browser window.  Open the above link manually to continue. ***")
		} else {
			log.Infoln(
				"Validation opened in a new browser window. ",
				"Please check your browser for further authentication steps.",
			)
		}
	} else {
		glog.Infoln("Please open the above URL in a browser and authenticate.")
	}

	// Step 3: Continue polling backend to see if request has been authenticated yet. Poll every 3 seconds for 2 minutes
	pollResponse, err := poll(challenge.PollingUrl)
	if err != nil { return err }

	jsonPollResponse, err := json.Marshal(pollResponse)
	if err != nil { return err }

	credentialsPath, err := getCredentialsPath()
	if err != nil { return err }
	err = ioutil.WriteFile(credentialsPath, jsonPollResponse, 0600)
	if err != nil { return err }
	return nil
}
