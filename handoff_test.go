package toystore

import (
	"testing"
	"time"
)

func TestHandoffPut(t *testing.T) {
	config := &Config{HandoffInterval: time.Millisecond * 10}
	client := NewRpcClient()
	h := NewHintedHandoff(config, client)

	h.Put("foo", "bar", "n1")
	h.Put("foo", "bar", "n1")
	h.Put("foo", "bar", "n2")

	if len(h.data) != 2 {
		t.Error("h should have 2 hints")
	}

	if len(h.data["n1"]) != 2 {
		t.Error("n1 should have 2 hints")
	}

	if len(h.data["n2"]) != 1 {
		t.Error("n1 should have 1 hints")
	}
}

func TestHandoffScan(t *testing.T) {
	config := &Config{HandoffInterval: time.Millisecond * 10}
	client := NewRpcClient()
	h := NewHintedHandoff(config, client)

	h.Put("foo", "bar", "n1")
	h.Put("foo", "bar", "n1")
	h.Put("foo", "bar", "n2")

	time.Sleep(config.HandoffInterval * 2)
}
