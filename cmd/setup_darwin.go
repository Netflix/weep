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
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

var (
	setupShortHelp = "Print setup information"
	setupLongHelp  = `By default, this command will print a script that can be used to set up IMDS routing.
If you trust us enough, you can let Weep do all the work for you:

sudo weep setup --commit

Otherwise, run weep setup and inspect the output. Then save the output to a file or pass it to your shell:

# Pass to shell
weep setup  # trust no one, always inspect
weep setup | sudo sh

# Save to file
weep setup > setup.sh
cat setup.sh  # trust no one, always inspect
chmod u+x setup.sh
./setup.sh`
	embedPrefix           = "extras/macos"
	pfRedirectionFilename = "/etc/pf.anchors/redirection"
	pfConfFilename        = "/etc/pf.conf"
	pfPlistFilename       = "/Library/LaunchDaemons/com.user.pfctl.plist"
	loopbackPlistFilename = "/Library/LaunchDaemons/com.user.lo0-loopback.plist"
)

// isRoot returns true if weep is running as root and false otherwise
func isRoot() bool {
	return os.Geteuid() == 0
}

// writeFileFromEmbedded calls the appropriate file writing function based on the value of commit
func writeFileFromEmbedded(prefix string, filename string, commit bool) error {
	data, err := SetupExtras.ReadFile(prefix + filename)
	port := viper.GetString("server.port")
	data = bytes.Replace(data, []byte("WEEP_PORT"), []byte(port), -1)
	if err != nil {
		return err
	}
	if commit {
		err = writeFileGo(filename, data)
		return err
	} else {
		err = writeFileShell(filename, data)
		return err
	}
}

// writeFileGo creates a backup of the target file and writes new content using Go
func writeFileGo(filename string, data []byte) error {
	print("writing ", filename, "...\n")
	// Ignore error on backup, we're just trying to be nice anyway
	_ = backupFile(filename)
	err := ioutil.WriteFile(filename, data, 0644)
	return err
}

// backupFile creates a copy of filename with a timestamp appended to the filename
func backupFile(filename string) error {
	src := filename
	dst := fmt.Sprintf("%s.%s", filename, time.Now().Format("20060102150405"))
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

// writeFileGo prints a script to create a backup of the target file and writes new content
func writeFileShell(filename string, data []byte) error {
	// make a backup copy, ignoring failure (e.g. if source file doesn't exist)
	fmt.Printf("cp %s %s.$(date +%%Y%%m%%d%%H%%M%%S) || true\n", filename, filename)
	fmt.Printf("cat << EOF > %s\n", filename)
	fmt.Print(string(data))
	fmt.Print("EOF\n")
	return nil
}

// reloadPlist uses executes launchctl to unload and load the specified plist, or prints the commands
// to do so if commit is false
func reloadPlist(plistFile string, commit bool) error {
	unloadCmd := exec.Command("launchctl", "unload", plistFile)
	loadCmd := exec.Command("launchctl", "load", plistFile)
	if commit {
		print("loading ", plistFile, "...\n")
		// Ignore error on unload because this plist might not exist
		_ = unloadCmd.Run()
		err := loadCmd.Run()
		return err
	}
	fmt.Println(unloadCmd.String(), " || true")
	fmt.Println(loadCmd.String())
	return nil
}

func Setup(cmd *cobra.Command, commit bool) error {
	if commit && !isRoot() {
		cmd.Print("not running as root. If this fails, try again using sudo.\n")
	}
	if !commit {
		fmt.Println("#!/bin/sh")
		fmt.Println("set -euo pipefail")
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
	if err := reloadPlist(pfPlistFilename, commit); err != nil {
		return err
	}
	if err := reloadPlist(loopbackPlistFilename, commit); err != nil {
		return err
	}
	return nil
}
