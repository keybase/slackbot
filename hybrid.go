package slackbot

import (
	"sync"
)

type hybridRunner struct {
	runner  BotCommandRunner
	channel string
}

func newHybridRunner(runner BotCommandRunner, channel string) *hybridRunner {
	return &hybridRunner{
		runner:  runner,
		channel: channel,
	}
}

func (r *hybridRunner) RunCommand(args []string, channel string) error {
	return r.runner.RunCommand(args, r.channel)

}

type HybridBackendMember struct {
	Backend BotBackend
	Channel string
}

type HybridBackend struct {
	backends []HybridBackendMember
}

func NewHybridBackend(backends ...HybridBackendMember) *HybridBackend {
	return &HybridBackend{
		backends: backends,
	}
}

func (b *HybridBackend) SendMessage(text string, channel string) {
	for _, backend := range b.backends {
		backend.Backend.SendMessage(text, backend.Channel)
	}
}

func (b *HybridBackend) Listen(runner BotCommandRunner) {
	var wg sync.WaitGroup
	for _, backend := range b.backends {
		wg.Add(1)
		go func() {
			backend.Backend.Listen(newHybridRunner(runner, backend.Channel))
			wg.Done()
		}()
	}
	wg.Wait()
}
