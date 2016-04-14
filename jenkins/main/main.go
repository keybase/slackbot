// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"flag"
	"log"

	"github.com/keybase/slackbot/jenkins"
)

func main() {
	var f struct {
		action             string
		clientRev          string
		kbfsRev            string
		jsonUpdateFilename string
	}
	flag.StringVar(&f.action, "action", "build", "Action")
	flag.StringVar(&f.clientRev, "client-rev", "", "Client commit")
	flag.StringVar(&f.kbfsRev, "kbfs-rev", "", "KBFS commit")
	flag.StringVar(&f.jsonUpdateFilename, "json", "", "JSON update filename")
	flag.Parse()

	switch f.action {
	case "stop":
		jenkins.StopBuild(flag.Arg(0))
	case "build":
		res, _ := jenkins.StartBuild(f.clientRev, f.kbfsRev, f.jsonUpdateFilename)
		log.Printf("Started: %s\n", res)
	}
}
