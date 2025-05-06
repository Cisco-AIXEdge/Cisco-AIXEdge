package main

import (
	"os"
	"strings"

	"github.com/iosxe-yosemite/IOS-XE-Copilot/internals"
)

func main() {
	//Client is initiated. Basically the struct defined in /internals/client.go
	//is initiated with values from /internals/meta.go
	client := internals.Client{}
	client.Init()
	//Entrypoint into the app
	if len(os.Args) == 1 {
		// If the app is called without arguments then client.Help()
		// is called which shows how the app can be launched (/internals/cli.go)
		client.Help()
		os.Exit(0)
	}
	if (os.Args[1] == "--prompt" || os.Args[1] == "-p") && len(os.Args) >= 3 {
		argsWithoutProg := os.Args[2:]
		justString := strings.Join(argsWithoutProg, " ")
		// Here are handled 2 cases where the users queries the copilot with an
		// output of a command or just a simple ask.
		if os.Args[2] == "show" {
			client.Prompt(justString)
		} else {
			client.Prompt(justString)
		}
	} else if os.Args[1] == "--interactive" || os.Args[1] == "-i" {
		client.Interactive()
	} else if os.Args[1] == "--upgrade" || os.Args[1] == "-u" {
		// The upgrade is triggered here. The upgrade function is in /internals/version.go
		client.CheckVersion()
	} else if os.Args[1] == "--config" || os.Args[1] == "-c" {
		// The configuration consists of writing in a JSON file the SN/PN and API key for AI engine
		// As argument it needs the API key. Check /internals/config.go
		client.ConfigWrite(os.Args[2])
	} else if os.Args[1] == "--version" || os.Args[1] == "-v" {
		// Shows the software version
		client.ShowVersion()
	} else if os.Args[1] == "--pcap" {
		// Shows the software version
		client.SetPcapFile(os.Args[2])

		argsWithoutProg := os.Args[3:]
		justString := strings.Join(argsWithoutProg, " ")
		client.Pcap(justString)

	} else if (os.Args[1] == "--optics" || os.Args[1] == "-o") && len(os.Args) >= 3 {
		// Shows the software version
		argsWithoutProg := os.Args[2:]
		justString := strings.Join(argsWithoutProg, " ")
		client.OpticsPrompt(justString)
	} else if (os.Args[1] == "--feature" || os.Args[1] == "-f") && len(os.Args) >= 3 {
		// Shows the software version
		argsWithoutProg := os.Args[2:]
		justString := strings.Join(argsWithoutProg, " ")
		client.FeaturePrompt(justString)

	} else if os.Args[1] == "--help" || os.Args[1] == "-h" {
		// If the app is calledwith --help or -h then client.Help()
		// is called which shows how the app can be launched (/internals/cli.go)
		client.Help()
	} else {
		client.Help()
	}

}
