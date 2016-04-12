package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/keybase/slackbot/jenkins"
)

func main() {
	command := "build"

	if len(os.Args) > 1 {
		command = strings.ToLower(os.Args[1])
	}

	switch command {
	case "stop":
		jenkins.StopBuild(os.Args[2])
	case "build":
		var clientRev, kbfsRev, JSONUpdateFilename string
		if len(os.Args) > 2 {
			clientRev = os.Args[2]
			if len(os.Args) > 3 {
				kbfsRev = os.Args[3]
				if len(os.Args) > 4 {
					JSONUpdateFilename = os.Args[4]
				}
			}
		}
		res, _ := jenkins.StartBuild(clientRev, kbfsRev, JSONUpdateFilename)
		fmt.Printf("Started: %s\n", res)
	}
}
