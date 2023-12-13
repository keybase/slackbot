module github.com/keybase/slackbot

go 1.19

require (
	github.com/keybase/go-keybase-chat-bot v0.0.0-20231213202706-1bdf50c0adf4
	github.com/nlopes/slack v0.1.1-0.20180101221843-107290b5bbaf
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
)

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20210927113745-59d0afb8317a // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	golang.org/x/net v0.0.0-20211118161319-6a13c67c3ce4 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// keybase maintained forks
replace (
	bazil.org/fuse => github.com/keybase/fuse v0.0.0-20210104232444-d36009698767
	bitbucket.org/ww/goautoneg => github.com/adjust/goautoneg v0.0.0-20150426214442-d788f35a0315
	github.com/stellar/go => github.com/keybase/stellar-org v0.0.0-20191010205648-0fc3bfe3dfa7
	github.com/syndtr/goleveldb => github.com/keybase/goleveldb v1.0.1-0.20211106225230-2a53fac0721c
	gopkg.in/src-d/go-billy.v4 => github.com/keybase/go-billy v3.1.1-0.20180828145748-b5a7b7bc2074+incompatible
	gopkg.in/src-d/go-git.v4 => github.com/keybase/go-git v4.0.0-rc9.0.20190209005256-3a78daa8ce8e+incompatible
	mvdan.cc/xurls/v2 => github.com/keybase/xurls/v2 v2.0.1-0.20190725180013-1e015cacd06c
)
