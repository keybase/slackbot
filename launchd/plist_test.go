// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package launchd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlist(t *testing.T) {
	env := NewEnv(os.Getenv("HOME"), "/usr/bin")
	data, err := env.Plist(Script{Label: "test.label", Path: "foo.sh", EnvVars: []EnvVar{{Key: "TEST", Value: "val"}}})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Plist: %s", string(data))
}

func TestWritePlistRestrictsExistingFilePermissions(t *testing.T) {
	home := t.TempDir()
	launchAgents := filepath.Join(home, "Library", "LaunchAgents")
	require.NoError(t, os.MkdirAll(launchAgents, 0o750))
	path := filepath.Join(launchAgents, "test.label.plist")
	//nolint:gosec // Deliberately reproduce the old insecure mode.
	require.NoError(t, os.WriteFile(path, []byte("old"), 0o755))

	env := NewEnv(home, "/usr/bin")
	writtenPath, err := env.WritePlist(Script{Label: "test.label", Path: "foo.sh"})
	require.NoError(t, err)
	require.Equal(t, path, writtenPath)
	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func TestValidateLabel(t *testing.T) {
	require.NoError(t, ValidateLabel("keybase.build.darwin"))
	for _, label := range []string{"", "--help", "../job", `..\job`} {
		require.Error(t, ValidateLabel(label), label)
	}
}
