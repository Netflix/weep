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
	"strings"
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
sudo ./setup.sh`
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

func embeddedFileData(prefix, filename string) ([]byte, error) {
	data, err := SetupExtras.ReadFile(prefix + filename)
	if err != nil {
		return nil, err
	}
	port := viper.GetString("server.port")
	data = bytes.Replace(data, []byte("WEEP_PORT"), []byte(port), -1)
	return data, nil
}

// writeFile creates a backup of the target file and writes new content using Go
func writeFile(prefix, filename string) error {
	print("writing ", filename, "...\n")
	data, err := embeddedFileData(prefix, filename)
	if err != nil {
		return err
	}
	// Ignore error on backup, we're just trying to be nice anyway
	_ = backupFile(filename)
	err = ioutil.WriteFile(filename, data, 0644)
	return err
}

// writeFileCommand returns commands to create a backup of the target file and write new content
func writeFileCommand(prefix, filename string) (string, error) {
	data, err := embeddedFileData(prefix, filename)
	if err != nil {
		return "", err
	}
	result := "printf \"backing up and overwriting " + filename + "...\"\n"
	result += fmt.Sprintf("cp %s %s.$(date +%%Y%%m%%d%%H%%M%%S) > /dev/null 2>&1 || true\n", filename, filename)
	result += fmt.Sprintf("cat << EOF > %s\n", filename)
	result += string(data)
	result += "EOF\n"
	result += "printf \" done\\n\"\n"
	return result, nil
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

	dest, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dest.Close()
	_, err = io.Copy(dest, source)
	return err
}

// reloadPlist executes launchctl to unload and load the specified plist
func reloadPlist(plistFile string) error {
	unloadCmd := exec.Command("launchctl", "unload", plistFile)
	loadCmd := exec.Command("launchctl", "load", plistFile)
	print("loading ", plistFile, "...\n")
	// Ignore error on unload because this plist might not exist
	_ = unloadCmd.Run()
	err := loadCmd.Run()
	return err
}

// reloadPlistCommand returns launchctl commands to unload and load the specified plist
func reloadPlistCommand(plistFile string) string {
	unloadCmd := exec.Command("launchctl", "unload", plistFile)
	loadCmd := exec.Command("launchctl", "load", plistFile)
	result := "printf \"reloading " + plistFile + "...\"\n"
	result += unloadCmd.String() + " > /dev/null 2>&1 || true"
	result += "\n"
	result += loadCmd.String()
	result += "\n"
	result += "printf \" done\\n\"\n"
	return result
}

func performSetup(files, plists []string) error {
	for _, file := range files {
		if err := writeFile(embedPrefix, file); err != nil {
			return err
		}
	}
	for _, plist := range plists {
		if err := reloadPlist(plist); err != nil {
			return err
		}
	}
	return nil
}

func buildScript(files, plists []string) (string, error) {
	var builder strings.Builder
	builder.WriteString("#!/bin/sh\n\n")
	for _, file := range files {
		if result, err := writeFileCommand(embedPrefix, file); err != nil {
			return "", err
		} else {
			builder.WriteString(result)
		}
		builder.WriteString("\n")
	}
	for _, plist := range plists {
		result := reloadPlistCommand(plist)
		builder.WriteString(result)
		builder.WriteString("\n")
	}
	return builder.String(), nil
}

func Setup(cmd *cobra.Command, commit bool) error {
	if commit && !isRoot() {
		cmd.Print("not running as root. If this fails, try again using sudo.\n")
	}
	files := []string{
		pfRedirectionFilename,
		pfConfFilename,
		pfPlistFilename,
		loopbackPlistFilename,
	}
	plists := []string{
		pfPlistFilename,
		loopbackPlistFilename,
	}

	if commit {
		if err := performSetup(files, plists); err != nil {
			return err
		}
	} else {
		if script, err := buildScript(files, plists); err != nil {
			return err
		} else {
			fmt.Println(script)
		}
	}
	return nil
}
