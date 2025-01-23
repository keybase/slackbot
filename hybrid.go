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

func (r *hybridRunner) RunCommand(args []string, _ string) error {
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

func (b *HybridBackend) SendMessage(text string, _ string) {
	for _, backend := range b.backends {
		backend.Backend.SendMessage(text, backend.Channel)
	}
}

func (b *HybridBackend) Listen(runner BotCommandRunner) {
	var wg sync.WaitGroup
	for _, backend := range b.backends {
		wg.Add(1)
		go func(b HybridBackendMember) {
			b.Backend.Listen(newHybridRunner(runner, b.Channel))
			wg.Done()
		}(backend)
	}
	wg.Wait()
}
