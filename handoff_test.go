package toystore

import (
	"testing"
	"time"

	"github.com/rlayte/toystore/data"
)

type FakeTransferrer struct {
	sent   map[string][]*data.Data
	status bool
}

func (f *FakeTransferrer) Transfer(address string, hints []*data.Data) bool {
	if _, ok := f.sent[address]; !ok {
		f.sent[address] = []*data.Data{}
	}

	f.sent[address] = append(f.sent[address], hints...)
	return true
}

func TestHandoffPut(t *testing.T) {
	config := Config{HandoffInterval: time.Millisecond * 10}
	client := &FakeTransferrer{map[string][]*data.Data{}, false}
	h := NewHintedHandoff(config, client)

	h.Put(data.New("foo", "bar"), "n1")
	h.Put(data.New("food", "bar"), "n1")
	h.Put(data.New("foo", "baz"), "n1")
	h.Put(data.New("food", "bar"), "n2")
	h.Put(data.New("foo", "baz"), "n2")

	// Client initially fails. Wait for it to recover and check hints are
	// still sent.
	time.Sleep(config.HandoffInterval * 2)
	client.status = true
	time.Sleep(config.HandoffInterval * 2)

	if len(client.sent["n1"]) != 3 {
		t.Errorf("Should have sent one item, but sent %d", len(client.sent["n1"]))
	}

	if len(client.sent["n2"]) != 2 {
		t.Errorf("Should have sent one item, but sent %d", len(client.sent["n1"]))
	}
}
