package toystore

import "testing"

func TestHandoffPut(t *testing.T) {
	c := &Config{}
	client := NewRpcClient()
	h := NewHintedHandoff(c, client)

	h.Put("foo", "bar", "n1")
	h.Put("foo", "bar", "n1")
	h.Put("foo", "bar", "n2")

	if len(h.data) != 2 {
		t.Error("h should have 2 hints")
	}
}
