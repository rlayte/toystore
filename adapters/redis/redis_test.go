package redis

import "testing"

func Equal(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Error("Not equal:", a, b)
	}
}

func NotEqual(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		t.Error("Not equal:", a, b)
	}
}

func TestRedisStore(t *testing.T) {
	res := NewRedisStore("localhost:6379")
	res.Put("foo", "bar")
	str, success := res.Get("foo")
	if !success {
		t.Error("Test Redis Store unsuccessful.")
	}
	Equal(t, str, "bar")
}

func TestFailure(t *testing.T) {
	res := NewRedisStore("localhost:6379")
	_, success := res.Get("lol")
	Equal(t, success, false)
}
