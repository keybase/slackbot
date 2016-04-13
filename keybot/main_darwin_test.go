// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build darwin

package main

import "testing"

func TestBuildPlease(t *testing.T) {
	out, err := kingpinHandler([]string{"build", "please"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "Dry Run: Doing that would run `/bin/launchctl` with args: [start keybase.prerelease]" {
		t.Errorf("Unexpected output: %s", out)
	}
}

func TestPromoteRelease(t *testing.T) {
	out, err := kingpinHandler([]string{"release", "promote", "1.2.3"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "Dry Run: Doing that would run `/bin/launchctl` with args: [start keybase.prerelease.promotearelease]" {
		t.Errorf("Unexpected output: %s", out)
	}
}
