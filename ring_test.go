package toystore

import (
	"container/list"
	"testing"
)

func init() {
	Hash = func(bytes []byte) []byte {
		return bytes
	}
}

func TestRinglessThan(t *testing.T) {
	cases := map[string]string{
		"a": "b",
		"c": "d",
		"f": "j",
	}

	for a, b := range cases {
		if lessThan(&list.Element{Value: a}, b) != true {
			t.Errorf("%s should be less than %s", a, b)
		}

		if lessThan(&list.Element{Value: b}, a) == true {
			t.Errorf("%s should not be less than %s", a, b)
		}
	}

	if lessThan(&list.Element{Value: "a"}, "a") != true {
		t.Errorf("The same value should return true")
	}
}
