package slackbot

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseRepositoryPathArgument(t *testing.T) {
	for input, expected := range map[string]string{
		"github.com/keybase/client":   "github.com/keybase/client",
		"`github.com/keybase/client`": "github.com/keybase/client",
		`github.com\keybase\client`:   "github.com/keybase/client",
	} {
		actual, err := ParseRepositoryArgument(input)
		require.NoError(t, err, input)
		require.Equal(t, expected, actual, input)
	}
	for _, input := range []string{"", ".", "..", "../../outside", "/tmp/repo", "$HOME/repo", `C:\repo`} {
		_, err := ParseRepositoryArgument(input)
		require.Error(t, err, input)
	}
}

func TestResolveRepositoryPath(t *testing.T) {
	root := t.TempDir()
	repository := filepath.Join(root, "github.com", "keybase", "client")
	require.NoError(t, os.MkdirAll(repository, 0o750))

	actual, err := ResolveRepositoryPath(root, "github.com/keybase/client")
	require.NoError(t, err)
	expected, err := filepath.EvalSymlinks(repository)
	require.NoError(t, err)
	require.Equal(t, expected, actual)

	if runtime.GOOS == "windows" {
		return
	}
	outside := t.TempDir()
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "outside-link")))
	_, err = ResolveRepositoryPath(root, "outside-link")
	require.Error(t, err)
}
