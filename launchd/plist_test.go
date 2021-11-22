// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package launchd

import (
	"os"
	"testing"
)

func TestPlist(t *testing.T) {
	env := NewEnv(os.Getenv("HOME"), "/usr/bin")
	data, err := env.Plist(Script{Label: "test.label", Path: "foo.sh", EnvVars: []EnvVar{{Key: "TEST", Value: "val"}}})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Plist: %s", string(data))
}
