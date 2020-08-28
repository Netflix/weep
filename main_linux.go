// +build linux

package main

import (
	"fmt"
)

func PrintSetup() {
	fmt.Println("Please run the following commands to setup routing for the meta-data service:")
	fmt.Println("")
	fmt.Println("\tsudo iptables -t nat -A OUTPUT -p tcp --dport 80 -d 169.254.169.254 -j DNAT --to 127.0.0.1:9090")
	fmt.Println("")
}
