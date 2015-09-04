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

func CheckInside(t *testing.T, strings []string, item string) {
	for _, str := range strings {
		if str == item {
			return
		}
	}
	t.Errorf("%s is not inside the given list.", item)
}

func TestRedisStore(t *testing.T) {
	res := New("localhost:6379")
	res.Put("foo", "bar")
	str, success := res.Get("foo")
	if !success {
		t.Error("Test Redis Store unsuccessful.")
	}
	Equal(t, str, "bar")
}

func TestFailure(t *testing.T) {
	res := New("localhost:6379")
	_, success := res.Get("lol")
	Equal(t, success, false)
}

func TestKeys(t *testing.T) {
	res := New("localhost:6379")
	res.Put("foo", "bar")
	res.Put("left", "right")
	if len(res.Keys()) == 0 {
		t.Error("Keys() returns an empty list")
	}
	CheckInside(t, res.Keys(), "foo")
	CheckInside(t, res.Keys(), "left")
}
