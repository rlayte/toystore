package toystore

import (
	"log"
	"time"

	"github.com/rlayte/toystore/data"
)

// HintedHandoff keeps track of data that should be stored on other
// nodes. It periodically scans this data and attempts to transfer
// it to the correct node. It then removes any data that is transferred.
type HintedHandoff struct {
	ScanInterval time.Duration

	data   map[string][]*data.Data
	client Transferrer
}

// scan periodically attempts to transfer hinted data to its correct
// location. If it removes any data it no longer needs.
func (h *HintedHandoff) scan() {
	for {
		for node, hints := range h.data {
			log.Println("Handoff", node, hints)
			ok := h.client.Transfer(node, hints)

			if ok {
				delete(h.data, node)
			}
		}

		time.Sleep(h.ScanInterval)
	}
}

// Put adds a new value for the hinted location.
func (h *HintedHandoff) Put(value *data.Data, hint string) {
	log.Println("Adding hint", value.Key, hint)
	if _, ok := h.data[hint]; !ok {
		h.data[hint] = []*data.Data{}
	}

	h.data[hint] = append(h.data[hint], value)
}

// NewHintedHandoff returns a new instance and starts the scan process
// using the HandoffInterval defined in config.
func NewHintedHandoff(config Config, client Transferrer) *HintedHandoff {
	h := &HintedHandoff{
		ScanInterval: config.HandoffInterval,
		data:         map[string][]*data.Data{},
		client:       client,
	}

	go h.scan()

	return h
}
