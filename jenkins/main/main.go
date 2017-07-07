// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"flag"
	"log"
	"time"

	"github.com/keybase/slackbot/jenkins"
)

func main() {
	var f struct {
		action        string
		clientRev     string
		kbfsRev       string
		updateChannel string
		crumb         string
	}
	flag.StringVar(&f.action, "action", "build", "Action")
	flag.StringVar(&f.clientRev, "client-rev", "", "Client commit")
	flag.StringVar(&f.kbfsRev, "kbfs-rev", "", "KBFS commit")
	flag.StringVar(&f.crumb, "crumb", "", "Jenkins Crumb")
	flag.StringVar(&f.updateChannel, "update-channel", "Test", "update channel: Test, Smoke, SmokeCI (default)")
	flag.Parse()

	if f.crumb != "" {
		jenkins.SetLastCrumb(jenkins.CrumbResult{Crumb: f.crumb, Time: time.Now()})
	}

	switch f.action {
	case "stop":
		jenkins.StopBuild(flag.Arg(0))
	case "build":
		res, _ := jenkins.StartBuild(f.clientRev, f.kbfsRev, f.updateChannel)
		log.Printf("Started: %s\n", res)
	}
	crumb := jenkins.GetLastCrumb()
	log.Printf("Crumb: %s\n", crumb.Crumb)
}
