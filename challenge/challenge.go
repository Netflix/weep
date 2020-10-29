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

package challenge

import (
	"encoding/json"
	"errors"
	"fmt"
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

	"github.com/spf13/viper"

	"github.com/golang/glog"
	"github.com/netflix/weep/util"
	log "github.com/sirupsen/logrus"
)

func NewHTTPClient(consolemeUrl string) (*http.Client, error) {
	if !HasValidJwt() {
		return nil, errors.New("Your authentication to ConsoleMe has expired. Please restart weep.")
	}
	var challenge ConsolemeChallengeResponse
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		return nil, err
	}
	credentialsPath, err := getCredentialsPath()
	if err != nil {
		return nil, err
	}
	challengeBody, err := ioutil.ReadFile(credentialsPath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(challengeBody, &challenge)
	if err != nil {
		return nil, err
	}
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
	if err != nil {
		return nil, err
	}
	jar.SetCookies(consoleMeUrlParsed, cookies)
	client := &http.Client{
		Jar: jar,
	}

	return client, err
}

func isWSL() bool {
	if util.FileExists("/proc/sys/kernel/osrelease") {
		if osrelease, err := ioutil.ReadFile("/proc/sys/kernel/osrelease"); err == nil {
			if strings.Contains(strings.ToLower(string(osrelease)), "microsoft") {
				return true
			}
		}
	}
	return false
}

func poll(pollingUrl string) (*ConsolemeChallengeResponse, error) {
	timeout := time.After(2 * time.Minute)
	tick := time.Tick(3 * time.Second)
	req, err := http.NewRequest("GET", pollingUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	// Keep trying until we're timed out or got a result or got an error
	for {
		select {
		case <-timeout:
			return nil, errors.New("*** Unable to validate Challenge Response after 2 minutes. Quitting. ***")
		case <-tick:
			pollResponse, err := pollRequest(client, req)
			if err != nil {
				return nil, err
			}
			if pollResponse.Status == "success" {
				return pollResponse, nil
			}
		}
	}
}

func pollRequest(c *http.Client, r *http.Request) (*ConsolemeChallengeResponse, error) {
	var pollResponse ConsolemeChallengeResponse
	var pollResponseBody []byte
	resp, err := c.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	pollResponseBody, err = ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(pollResponseBody, &pollResponse)
	return &pollResponse, err
}

func getCredentialsPath() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
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
	if viper.GetString("challenge_settings.user") == "" {
		log.Fatalf(
			"Invalid configuration. You must define challenge_settings.user as the user you wish to authenticate as.",
		)
	}
	var consoleMeChallengeGeneratorEndpoint = fmt.Sprintf(
		"%s/noauth/v1/challenge_generator/%s",
		viper.GetString("consoleme_url"),
		viper.GetString("challenge_settings.user"),
	)
	var challenge ConsolemeChallenge
	req, err := http.NewRequest("GET", consoleMeChallengeGeneratorEndpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	tokenResponseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

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
	if err != nil {
		return err
	}

	jsonPollResponse, err := json.Marshal(pollResponse)
	if err != nil {
		return err
	}

	credentialsPath, err := getCredentialsPath()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(credentialsPath, jsonPollResponse, 0600)
	if err != nil {
		return err
	}
	return nil
}
