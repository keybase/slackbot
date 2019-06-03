// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package launchd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Env is environment for launchd
type Env struct {
	Path              string
	Home              string
	GoPath            string
	GoPathForBot      string
	GithubToken       string
	SlackToken        string
	SlackChannel      string
	AWSAccessKey      string
	AWSSecretKey      string
	KeybaseToken      string
	KeybaseChatConvID string
	KeybaseLocation   string
	KeybaseHome       string
}

// Script is what to run
type Script struct {
	Label      string
	Path       string
	BucketName string
	Platform   string
	LogPath    string
	EnvVars    []EnvVar
}

// EnvVar is custom env vars
type EnvVar struct {
	Key   string
	Value string
}

type job struct {
	Env     Env
	Script  Script
	LogPath string
}

const plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{{.Script.Label }}</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>GOPATH</key>
        <string>{{ .Env.GoPath }}</string>
        <key>GITHUB_TOKEN</key>
        <string>{{ .Env.GithubToken }}</string>
        <key>SLACK_TOKEN</key>
        <string>{{ .Env.SlackToken }}</string>
        <key>SLACK_CHANNEL</key>
        <string>{{ .Env.SlackChannel }}</string>
        <key>AWS_ACCESS_KEY</key>
        <string>{{ .Env.AWSAccessKey }}</string>
        <key>AWS_SECRET_KEY</key>
        <string>{{ .Env.AWSSecretKey }}</string>
        <key>KEYBASE_TOKEN</key>
        <string>{{ .Env.KeybaseToken }}</string>
        <key>KEYBASE_CHAT_CONVID</key>
        <string>{{ .Env.KeybaseChatConvID }}</string>
        <key>KEYBASE_LOCATION</key>
        <string>{{ .Env.KeybaseLocation }}</string>
        <key>KEYBASE_HOME</key>
		<string>{{ .Env.KeybaseHome }}</string>
		<key>KEYBASE_RUN_MODE</key>
        <string>prod</string>
        <key>PATH</key>
        <string>{{ .Env.Path }}</string>
        <key>LOG_PATH</key>
        <string>{{ .Env.Home }}/Library/Logs/{{ .Script.Label }}.log</string>
        <key>BUCKET_NAME</key>
        <string>{{ .Script.BucketName }}</string>
        <key>SCRIPT_PATH</key>
        <string>{{ .Env.GoPath }}/src/{{ .Script.Path }}</string>
        <key>PLATFORM</key>
        <string>{{ .Script.Platform }}</string>
        <key>LABEL</key>
        <string>{{ .Script.Label }}</string>
        {{ with .Script.EnvVars }}{{ range . }}
        <key>{{ .Key }}</key>
        <string>{{ .Value }}</string>
        {{ end }}{{ end }}
    </dict>
    <key>ProgramArguments</key>
    <array>
        <string>/bin/bash</string>
        <string>{{ .Env.GoPathForBot }}/src/github.com/keybase/slackbot/scripts/run.sh</string>
    </array>
    <key>StandardErrorPath</key>
    <string>{{ .LogPath }}</string>
    <key>StandardOutPath</key>
    <string>{{ .LogPath }}</string>
</dict>
</plist>
`

// NewEnv creates environment
func NewEnv(home string, path string) Env {
	return Env{
		Path:              path,
		Home:              home,
		GoPath:            os.Getenv("GOPATH"),
		GoPathForBot:      os.Getenv("GOPATH"),
		GithubToken:       os.Getenv("GITHUB_TOKEN"),
		SlackToken:        os.Getenv("SLACK_TOKEN"),
		SlackChannel:      os.Getenv("SLACK_CHANNEL"),
		AWSAccessKey:      os.Getenv("AWS_ACCESS_KEY"),
		AWSSecretKey:      os.Getenv("AWS_SECRET_KEY"),
		KeybaseToken:      os.Getenv("KEYBASE_TOKEN"),
		KeybaseChatConvID: os.Getenv("KEYBASE_CHAT_CONVID"),
		KeybaseHome:       os.Getenv("KEYBASE_HOME"),
		KeybaseLocation:   os.Getenv("KEYBASE_LOCATION"),
	}
}

// PathFromHome returns path from home dir for env
func (e Env) PathFromHome(path string) string {
	return filepath.Join(os.Getenv("HOME"), path)
}

// LogPathForLaunchdLabel returns path to log for label
func (e Env) LogPathForLaunchdLabel(label string) (string, error) {
	if strings.Contains(label, "..") || strings.Contains(label, "/") || strings.Contains(label, `\`) {
		return "", fmt.Errorf("Invalid label")
	}
	return filepath.Join(e.Home, "Library/Logs", label+".log"), nil
}

// Plist is plist for env and args
func (e Env) Plist(script Script) ([]byte, error) {
	t := template.New("Plist template")
	logPath, lerr := e.LogPathForLaunchdLabel(script.Label)
	if lerr != nil {
		return nil, lerr
	}
	j := job{Env: e, Script: script, LogPath: logPath}
	t, err := t.Parse(plistTemplate)
	if err != nil {
		return nil, err
	}
	buff := bytes.NewBufferString("")
	err = t.Execute(buff, j)
	if err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

// WritePlist writes out plist and returns path that was written to
func (e Env) WritePlist(script Script) (string, error) {
	data, err := e.Plist(script)
	if err != nil {
		return "", err
	}
	plistDir := e.Home + "/Library/LaunchAgents"
	if err := os.MkdirAll(plistDir, 0755); err != nil {
		return "", err
	}
	path := fmt.Sprintf("%s/%s.plist", plistDir, script.Label)
	log.Printf("Writing %s", path)
	if err := ioutil.WriteFile(path, data, 0755); err != nil {
		return "", err
	}
	return path, nil
}

// Cleanup removes any files generated by Env
func (e Env) Cleanup(script Script) error {
	plistDir := e.Home + "/Library/LaunchAgents"
	path := fmt.Sprintf("%s/%s.plist", plistDir, script.Label)
	log.Printf("Removing %s", path)
	return os.Remove(path)
}

// CleanupLog removes log path
func CleanupLog(env Env, label string) error {
	// Remove log
	logPath, lerr := env.LogPathForLaunchdLabel(label)
	if lerr != nil {
		return lerr
	}
	if _, err := os.Stat(logPath); err == nil {
		if err := os.Remove(logPath); err != nil {
			return err
		}
	}
	return nil
}
