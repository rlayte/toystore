package memory

import "testing"

func Equal(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Error("Not equal:", a, b)
	}
}

func TestMemoryStore(t *testing.T) {
	res := New()
	res.Put("foo", "bar")
	str, success := res.Get("foo")
	if !success {
		t.Error("Test Memory Store unsuccessful.")
	}
	Equal(t, str, "bar")
}

func TestFailure(t *testing.T) {
	res := New()
	_, success := res.Get("lol")
	Equal(t, success, false)
}
