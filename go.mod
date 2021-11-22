module github.com/keybase/slackbot

go 1.17

require (
	github.com/keybase/client/go v0.0.0-20211122170721-a66861b2e521
	github.com/keybase/go-keybase-chat-bot v0.0.0-20211119193246-0a6a7b508a0e
	github.com/nlopes/slack v0.1.1-0.20180101221843-107290b5bbaf
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
)

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20210927113745-59d0afb8317a // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/keybase/backoff v1.0.1-0.20160517061000-726b63b835ec // indirect
	github.com/keybase/clockwork v0.1.1-0.20161209210251-976f45f4a979 // indirect
	github.com/keybase/go-codec v0.0.0-20180928230036-164397562123 // indirect
	github.com/keybase/go-framed-msgpack-rpc v0.0.0-20211118173254-f892386581e8 // indirect
	github.com/keybase/go-jsonw v0.0.0-20200131153605-3e5b58caddd9 // indirect
	github.com/keybase/msgpackzip v0.0.0-20211109205514-10e4bc329851 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/testify v1.7.0 // indirect
	golang.org/x/net v0.0.0-20211118161319-6a13c67c3ce4 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
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
