package toystore

import (
	"time"

	"github.com/rlayte/toystore/data"
)

type HintedHandoff struct {
	ScanInterval time.Duration

	data   map[string]*data.Data
	client PeerClient
}

func (h *HintedHandoff) scan() {
	for {
		time.Sleep(h.ScanInterval)
	}
}

func (h *HintedHandoff) Put(key string, value interface{}, hint string) {
	h.data[hint] = data.NewData(key, value)
}

func NewHintedHandoff(config *Config, client PeerClient) *HintedHandoff {
	h := &HintedHandoff{
		ScanInterval: config.HandoffInterval,
		data:         map[string]*data.Data{},
		client:       client,
	}

	go h.scan()

	return h
}
