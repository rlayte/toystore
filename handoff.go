package toystore

import (
	"time"

	"github.com/rlayte/toystore/data"
)

type HintedHandoff struct {
	ScanInterval time.Duration

	data   map[string][]*data.Data
	client PeerClient
}

func (h *HintedHandoff) scan() {
	for {
		for node, hints := range h.data {
			ok := h.client.Transfer(node, hints)

			if ok {
				delete(h.data, node)
			}
		}

		time.Sleep(h.ScanInterval)
	}
}

func (h *HintedHandoff) Put(value *data.Data, hint string) {
	if _, ok := h.data[hint]; !ok {
		h.data[hint] = []*data.Data{}
	}

	h.data[hint] = append(h.data[hint], value)
}

func NewHintedHandoff(config Config, client PeerClient) *HintedHandoff {
	h := &HintedHandoff{
		ScanInterval: config.HandoffInterval,
		data:         map[string][]*data.Data{},
		client:       client,
	}

	go h.scan()

	return h
}
