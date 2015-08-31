package toystore

import "time"

type HintedHandoff struct {
	ScanInterval time.Duration

	data   map[string]*Data
	client PeerClient
}

func (h *HintedHandoff) scan() {
	for {
		time.Sleep(h.ScanInterval)
	}
}

func (h *HintedHandoff) Put(key string, value interface{}, hint string) {
	h.data[hint] = NewData(key, value)
}

func NewHintedHandoff(config *Config, client PeerClient) *HintedHandoff {
	h := &HintedHandoff{
		ScanInterval: config.HandoffInterval,
		data:         map[string]*Data{},
		client:       client,
	}

	go h.scan()

	return h
}
