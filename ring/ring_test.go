package ring

import (
	"container/list"
	"fmt"
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

func TestRingAdd(t *testing.T) {
	ring := NewHashRing()
	ring.Add("d")
	ring.Add("c")
	ring.Add("e")
	ring.Add("b")
	ring.Add("a")

	if fmt.Sprint(ring) != "a, b, c, d, e" {
		t.Errorf("%s != 'a, b, c, d, e'", ring)
	}
}

func TestRingFind(t *testing.T) {
	ring := NewHashRing()
	ring.Add("d")
	ring.Add("e")
	ring.Add("b")
	ring.Add("a")

	if ring.Find("c") != "d" {
		t.Error("c is located on d")
	}

	if ring.Find("b") != "d" {
		t.Error("b is located on d, not", ring.Find("b"))
	}

	if ring.Find("1") != "a" {
		t.Error("f is located on a")
	}

	if ring.Find("f") != "a" {
		t.Error("f is located on a")
	}
}

func TestRingFindN(t *testing.T) {
	ring := NewHashRing()
	ring.Add("d")
	ring.Add("e")
	ring.Add("b")
	ring.Add("a")

	nodes := ring.FindN("c", 3)

	if len(nodes) != 3 {
		t.Error("FindN should return N keys")
	}

	for _, key := range []string{"d", "e", "a"} {
		if _, ok := nodes[key]; !ok {
			t.Errorf("Nodes should contain %s", key)
		}
	}
}

func TestRingFindNWithFailures(t *testing.T) {
	ring := NewHashRing()
	ring.Add("d")
	ring.Add("e")
	ring.Add("b")
	ring.Add("a")

	ring.Fail("e")

	nodes := ring.FindN("c", 3)

	if len(nodes) != 3 {
		t.Error("FindN should return N keys")
	}

	if nodes["b"] != "e" {
		t.Errorf("Hint should have been d, but was %s", nodes["b"])
	}

	for _, key := range []string{"d", "a", "b"} {
		if _, ok := nodes[key]; !ok {
			t.Errorf("Nodes should contain %s", key)
		}
	}
}

func TestRingAdjacent(t *testing.T) {
	ring := NewHashRing()
	ring.Add("d")
	ring.Add("e")
	ring.Add("b")
	ring.Add("a")

	if ring.Adjacent("c", "b") != true {
		t.Error("a should be next to b:", ring)
	}

	if ring.Adjacent("a", "b") != true {
		t.Error("a should be next to b:", ring)
	}

	if ring.Adjacent("e", "a") != true {
		t.Error("e should be next to a:", ring)
	}

	if ring.Adjacent("a", "d") == true {
		t.Error("a should not be next to d:", ring)
	}
}
