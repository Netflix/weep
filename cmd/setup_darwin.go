// +build darwin

package cmd

import (
	"fmt"
)

func PrintSetup() {
	fmt.Println("Please run the following commands to setup routing for the meta-data service:")
	fmt.Println("")
	fmt.Println("\tsudo ifconfig lo0 169.254.169.254 alias")
	fmt.Println("\techo \"rdr pass on lo0 inet proto tcp from any to 169.254.169.254 port 80 -> 127.0.0.1 port 9090\" | sudo pfctl -ef -")
	fmt.Println("")
}
