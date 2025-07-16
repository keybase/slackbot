// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/keybase/slackbot"
	"github.com/keybase/slackbot/cli"
	"github.com/keybase/slackbot/launchd"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type keybot struct{}

func (k *keybot) Run(bot *slackbot.Bot, channel string, args []string) (string, error) {
	app := kingpin.New("keybot", "Job command parser for keybot")
	app.Terminate(nil)
	stringBuffer := new(bytes.Buffer)
	app.Writer(stringBuffer)

	build := app.Command("build", "Build things")

	cancel := app.Command("cancel", "Cancel")
	cancelLabel := cancel.Arg("label", "Launchd job label").String()

	buildMobile := build.Command("mobile", "Start an iOS and Android build")
	buildMobileSkipCI := buildMobile.Flag("skip-ci", "Whether to skip CI").Bool()
	buildMobileAutomated := buildMobile.Flag("automated", "Whether this is a timed build").Bool()
	buildMobileCientCommit := buildMobile.Flag("client-commit", "Build a specific client commit hash").String()

	buildAndroid := build.Command("android", "Start an android build")
	buildAndroidSkipCI := buildAndroid.Flag("skip-ci", "Whether to skip CI").Bool()
	buildAndroidAutomated := buildAndroid.Flag("automated", "Whether this is a timed build").Bool()
	buildAndroidCientCommit := buildAndroid.Flag("client-commit", "Build a specific client commit hash").String()
	buildIOS := build.Command("ios", "Start an ios build")
	buildIOSClean := buildIOS.Flag("clean", "Whether to clean first").Bool()
	buildIOSSkipCI := buildIOS.Flag("skip-ci", "Whether to skip CI").Bool()
	buildIOSAutomated := buildIOS.Flag("automated", "Whether this is a timed build").Bool()
	buildIOSCientCommit := buildIOS.Flag("client-commit", "Build a specific client commit hash").String()

	buildDarwin := build.Command("darwin", "Start a darwin build")
	buildDarwinTest := buildDarwin.Flag("test", "Whether build is for testing").Bool()
	buildDarwinClientCommit := buildDarwin.Flag("client-commit", "Build a specific client commit").String()
	buildDarwinKbfsCommit := buildDarwin.Flag("kbfs-commit", "Build a specific kbfs commit").String()
	buildDarwinNoPull := buildDarwin.Flag("skip-pull", "Don't pull before building the app").Bool()
	buildDarwinSkipCI := buildDarwin.Flag("skip-ci", "Whether to skip CI").Bool()
	buildDarwinSmoke := buildDarwin.Flag("smoke", "Whether to make a pair of builds for smoketesting when on a branch").Bool()
	buildDarwinNoS3 := buildDarwin.Flag("skip-s3", "Don't push to S3 after building the app").Bool()
	buildDarwinNoNotarize := buildDarwin.Flag("skip-notarize", "Don't notarize the app").Bool()

	release := app.Command("release", "Release things")
	releasePromote := release.Command("promote", "Promote a release to public")
	releaseToPromotePlatform := releasePromote.Arg("platform", "Platform to promote a release for").Required().String()
	releaseToPromote := releasePromote.Arg("release-to-promote", "Promote a specific release to public immediately").Required().String()
	releaseToPromoteDryRun := releasePromote.Flag("dry-run", "Announce what would be done without doing it").Bool()

	releaseBroken := release.Command("broken", "Mark a release as broken")
	releaseBrokenVersion := releaseBroken.Arg("version", "Mark a release as broken").Required().String()

	smoketest := app.Command("smoketest", "Set the smoke testing status of a build")
	smoketestBuildA := smoketest.Flag("build-a", "The first of the two IDs comprising the new build").Required().String()
	smoketestPlatform := smoketest.Flag("platform", "The build's platform (darwin, linux, windows)").Required().String()
	smoketestEnable := smoketest.Flag("enable", "Whether smoketesting should be enabled").Required().Bool()
	smoketestMaxTesters := smoketest.Flag("max-testers", "Max number of testers for this build").Required().Int()

	dumplogCmd := app.Command("dumplog", "Show the log file")
	dumplogCommandLabel := dumplogCmd.Arg("label", "Launchd job label").Required().String()

	gitDiffCmd := app.Command("gdiff", "Show the git diff")
	gitDiffRepo := gitDiffCmd.Arg("repo", "Repo path relative to $GOPATH/src").Required().String()

	gitCleanCmd := app.Command("gclean", "Clean the repos go/go-ios/go-android")
	nodeModuleCleanCmd := app.Command("nodeModuleClean", "Clean the ios/android node_modules")

	upgrade := app.Command("upgrade", "Upgrade package")
	upgradePackageName := upgrade.Arg("name", "Package name (yarn, go, fastlane, etc)").Required().String()

	cmd, usage, cmdErr := cli.Parse(app, args, stringBuffer)
	if usage != "" || cmdErr != nil {
		return usage, cmdErr
	}

	home := os.Getenv("HOME")
	javaHome := "/Library/Java/JavaVirtualMachines/zulu-17.jdk/Contents/Home"
	path := javaHome  + "/bin:" + "/sbin:/usr/sbin:/bin:/usr/local/bin:/usr/bin:/opt/homebrew/bin"
	env := launchd.NewEnv(home, path)
	androidHome := "/usr/local/opt/android-sdk"
	ndkVer65x := "23.1.7779620"
	// ndkVer66x := "26.1.10909125"
	ndkVer := ndkVer65x 
	NDKPath := "/Users/build/Library/Android/sdk/ndk/" + ndkVer

	switch cmd {
	case cancel.FullCommand():
		if *cancelLabel == "" {
			return "Label required for cancel", errors.New("Label required for cancel")
		}
		return launchd.Stop(*cancelLabel)

	case buildDarwin.FullCommand():
		smokeTest := true
		skipCI := *buildDarwinSkipCI
		testBuild := *buildDarwinTest
		// If it's a custom build, make it a test build unless --smoke is passed.
		if *buildDarwinClientCommit != "" || *buildDarwinKbfsCommit != "" {
			smokeTest = *buildDarwinSmoke
			testBuild = !*buildDarwinSmoke
		}
		script := launchd.Script{
			Label:      "keybase.build.darwin",
			Path:       "github.com/keybase/client/packaging/build_darwin.sh",
			BucketName: "prerelease.keybase.io",
			Platform:   "darwin",
			EnvVars: []launchd.EnvVar{
				{Key: "SMOKE_TEST", Value: boolToEnvString(smokeTest)},
				{Key: "TEST", Value: boolToEnvString(testBuild)},
				{Key: "CLIENT_COMMIT", Value: *buildDarwinClientCommit},
				{Key: "KBFS_COMMIT", Value: *buildDarwinKbfsCommit},
				// TODO: Rename to SKIP_CI in packaging scripts
				{Key: "NOWAIT", Value: boolToEnvString(skipCI)},
				{Key: "NOPULL", Value: boolToEnvString(*buildDarwinNoPull)},
				{Key: "NOS3", Value: boolToEnvString(*buildDarwinNoS3)},
				{Key: "NONOTARIZE", Value: boolToEnvString(*buildDarwinNoNotarize)},
			},
		}
		return runScript(bot, channel, env, script)

	case buildMobile.FullCommand():
		skipCI := *buildMobileSkipCI
		automated := *buildMobileAutomated
		script := launchd.Script{
			Label:      "keybase.build.mobile",
			Path:       "github.com/keybase/client/packaging/build_mobile.sh",
			BucketName: "prerelease.keybase.io",
			EnvVars: []launchd.EnvVar{
				{Key: "ANDROID_HOME", Value: androidHome},
				{Key: "ANDROID_SDK", Value: androidHome},
				{Key: "ANDROID_SDK_ROOT", Value: androidHome},
				{Key: "ANDROID_NDK_HOME", Value: NDKPath},
				{Key: "NDK_HOME", Value: NDKPath},
				{Key: "ANDROID_NDK", Value: NDKPath},
				{Key: "CLIENT_COMMIT", Value: *buildMobileCientCommit},
				{Key: "CHECK_CI", Value: boolToEnvString(!skipCI)},
				{Key: "AUTOMATED_BUILD", Value: boolToEnvString(automated)},
			},
		}
		env.GoPath = env.PathFromHome("go-ios")
		return runScript(bot, channel, env, script)

	case buildAndroid.FullCommand():
		skipCI := *buildAndroidSkipCI
		automated := *buildAndroidAutomated
		script := launchd.Script{
			Label:      "keybase.build.android",
			Path:       "github.com/keybase/client/packaging/android/build_and_publish.sh",
			BucketName: "prerelease.keybase.io",
			EnvVars: []launchd.EnvVar{
				{Key: "ANDROID_HOME", Value: androidHome},
				{Key: "ANDROID_NDK_HOME", Value: NDKPath},
				{Key: "ANDROID_NDK", Value: NDKPath},
				{Key: "CLIENT_COMMIT", Value: *buildAndroidCientCommit},
				{Key: "CHECK_CI", Value: boolToEnvString(!skipCI)},
				{Key: "AUTOMATED_BUILD", Value: boolToEnvString(automated)},
			},
		}
		env.GoPath = env.PathFromHome("go-android") // Custom go path for Android so we don't conflict
		return runScript(bot, channel, env, script)

	case buildIOS.FullCommand():
		skipCI := *buildIOSSkipCI
		iosClean := *buildIOSClean
		automated := *buildIOSAutomated
		script := launchd.Script{
			Label:      "keybase.build.ios",
			Path:       "github.com/keybase/client/packaging/ios/build_and_publish.sh",
			BucketName: "prerelease.keybase.io",
			EnvVars: []launchd.EnvVar{
				{Key: "CLIENT_COMMIT", Value: *buildIOSCientCommit},
				{Key: "CLEAN", Value: boolToEnvString(iosClean)},
				{Key: "CHECK_CI", Value: boolToEnvString(!skipCI)},
				{Key: "AUTOMATED_BUILD", Value: boolToEnvString(automated)},
			},
		}
		env.GoPath = env.PathFromHome("go-ios") // Custom go path for iOS so we don't conflict
		return runScript(bot, channel, env, script)

	case releasePromote.FullCommand():
		script := launchd.Script{
			Label:      "keybase.release.promote",
			Path:       "github.com/keybase/slackbot/scripts/release.promote.sh",
			BucketName: "prerelease.keybase.io",
			Platform:   *releaseToPromotePlatform,
			EnvVars: []launchd.EnvVar{
				{Key: "RELEASE_TO_PROMOTE", Value: *releaseToPromote},
				{Key: "DRY_RUN", Value: boolToString(*releaseToPromoteDryRun)},
			},
		}
		return runScript(bot, channel, env, script)

	case dumplogCmd.FullCommand():
		readPath, err := env.LogPathForLaunchdLabel(*dumplogCommandLabel)
		if err != nil {
			return "", err
		}
		script := launchd.Script{
			Label:      "keybase.dumplog",
			Path:       "github.com/keybase/slackbot/scripts/dumplog.sh",
			BucketName: "prerelease.keybase.io",
			EnvVars: []launchd.EnvVar{
				{Key: "READ_PATH", Value: readPath},
				{Key: "NOLOG", Value: boolToEnvString(true)},
			},
		}
		return runScript(bot, channel, env, script)

	case gitDiffCmd.FullCommand():
		rawRepoText := *gitDiffRepo
		repoParsed := strings.Split(strings.Trim(rawRepoText, "`<>"), "|")[1]

		script := launchd.Script{
			Label:      "keybase.gitdiff",
			Path:       "github.com/keybase/slackbot/scripts/run_and_send_stdout.sh",
			BucketName: "prerelease.keybase.io",
			EnvVars: []launchd.EnvVar{
				{Key: "REPO", Value: repoParsed},
				{Key: "PREFIX_GOPATH", Value: boolToEnvString(true)},
				{Key: "SCRIPT_TO_RUN", Value: "./git_diff.sh"},
			},
		}
		return runScript(bot, channel, env, script)

	case gitCleanCmd.FullCommand():
		script := launchd.Script{
			Label:      "keybase.gitclean",
			Path:       "github.com/keybase/slackbot/scripts/run_and_send_stdout.sh",
			BucketName: "prerelease.keybase.io",
			EnvVars: []launchd.EnvVar{
				{Key: "SCRIPT_TO_RUN", Value: "./git_clean.sh"},
			},
		}
		return runScript(bot, channel, env, script)

	case nodeModuleCleanCmd.FullCommand():
		script := launchd.Script{
			Label:      "keybase.nodeModuleClean",
			Path:       "github.com/keybase/slackbot/scripts/run_and_send_stdout.sh",
			BucketName: "prerelease.keybase.io",
			EnvVars: []launchd.EnvVar{
				{Key: "SCRIPT_TO_RUN", Value: "./node_module_clean.sh"},
			},
		}
		return runScript(bot, channel, env, script)

	case releaseBroken.FullCommand():
		script := launchd.Script{
			Label:      "keybase.release.broken",
			Path:       "github.com/keybase/slackbot/scripts/release.broken.sh",
			BucketName: "prerelease.keybase.io",
			Platform:   "darwin",
			EnvVars: []launchd.EnvVar{
				{Key: "BROKEN_RELEASE", Value: *releaseBrokenVersion},
			},
		}
		return runScript(bot, channel, env, script)

	case smoketest.FullCommand():
		script := launchd.Script{
			Label:      "keybase.smoketest",
			Path:       "github.com/keybase/slackbot/scripts/smoketest.sh",
			BucketName: "prerelease.keybase.io",
			Platform:   *smoketestPlatform,
			EnvVars: []launchd.EnvVar{
				{Key: "SMOKETEST_BUILD_A", Value: *smoketestBuildA},
				{Key: "SMOKETEST_MAX_TESTERS", Value: strconv.Itoa(*smoketestMaxTesters)},
				{Key: "SMOKETEST_ENABLE", Value: boolToString(*smoketestEnable)},
			},
		}
		return runScript(bot, channel, env, script)

	case upgrade.FullCommand():
		script := launchd.Script{
			Label: "keybase.update",
			Path:  "github.com/keybase/slackbot/scripts/upgrade.sh",
			EnvVars: []launchd.EnvVar{
				{Key: "NAME", Value: *upgradePackageName},
			},
		}
		return runScript(bot, channel, env, script)
	}

	return cmd, nil
}

func (k *keybot) Help(bot *slackbot.Bot) string {
	out, err := k.Run(bot, "", nil)
	if err != nil {
		return fmt.Sprintf("Error getting help: %s", err)
	}
	return out
}
