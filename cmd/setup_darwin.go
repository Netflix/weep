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

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var setupShortHelp = "Print setup information"
var setupLongHelp = `By default, this command will print a script that can be used to set up IMDS routing.
If you trust us enough, you can run sudo weep setup --write.
Otherwise, run weep setup and inspect the output. Then run sudo eval $(weep setup).`

var (
	embedPrefix           = "extras/macos"
	pfRedirectionFilename = "/etc/pf.anchors/redirection"
	pfConfFilename        = "/etc/pf.conf"
	pfPlistFilename       = "/Library/LaunchDaemons/com.user.pfctl.plist"
	loopbackPlistFilename = "/Library/LaunchDaemons/com.user.lo0-loopback.plist"
)

func PrintSetup(cmd *cobra.Command) {
	cmd.Println("Please run the following commands to setup routing for the meta-data service:")
	cmd.Println("sudo ifconfig lo0 169.254.169.254 alias")
	cmd.Println("echo \"rdr pass on lo0 inet proto tcp from any to 169.254.169.254 port 80 -> 127.0.0.1 port 9091\" | sudo pfctl -ef -")
}

func isRoot() bool {
	return os.Geteuid() == 0
}

func writeFileFromEmbedded(prefix string, filename string, commit bool) error {
	data, err := SetupExtras.ReadFile(prefix + filename)
	if err != nil {
		return err
	}
	if commit {
		print("writing ", filename, "...\n")
		err = ioutil.WriteFile(filename, data, 0644)
		return err
	}
	fmt.Printf("cat << EOF > %s\n", filename)
	fmt.Print(string(data))
	fmt.Print("EOF\n")
	return nil
}

func loadPlist(plistFile string, commit bool) error {
	cmd := exec.Command("launchctl", "load", plistFile)
	if commit {
		print("loading ", plistFile, "...\n")
		err := cmd.Run()
		return err
	}
	fmt.Println(cmd.String())
	return nil
}

func Setup(cmd *cobra.Command, commit bool) error {
	if commit && !isRoot() {
		cmd.Print("not running as root. If this fails, try again using sudo.\n")
	}
	if !commit {
		fmt.Println("#!/usr/bin/env bash")
	}
	// copy redirection file to pf.anchors
	if err := writeFileFromEmbedded(embedPrefix, pfRedirectionFilename, commit); err != nil {
		return err
	}
	// replace pf.conf with ours
	if err := writeFileFromEmbedded(embedPrefix, pfConfFilename, commit); err != nil {
		return err
	}
	// copy plist files to launchdaemons
	if err := writeFileFromEmbedded(embedPrefix, pfPlistFilename, commit); err != nil {
		return err
	}
	if err := writeFileFromEmbedded(embedPrefix, loopbackPlistFilename, commit); err != nil {
		return err
	}
	// load plist files with launchctl
	if err := loadPlist(pfPlistFilename, commit); err != nil {
		return err
	}
	if err := loadPlist(loopbackPlistFilename, commit); err != nil {
		return err
	}
	return nil
}
