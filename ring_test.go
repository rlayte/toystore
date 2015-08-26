package toystore

import (
	"bytes"
	"testing"
)

func init() {
	Hash = func(bytes []byte) []byte {
		return bytes
	}
}

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

func TestAdd(t *testing.T) {
	a := NewRingHead()
	b := a.Add(NewRingString("b"))
	c := a.AddString("c")
	Equal(t, a, a)
	NotEqual(t, a, b)
	NotEqual(t, c, b)
	Equal(t, a.next, b)
	Equal(t, b.next, c)
	Equal(t, c.next, a)
}

func TestNode(t *testing.T) {
	var val []byte
	var err error
	c := RingFromList([]string{"1", "3", "5"})
	val, err = c.KeyAddress([]byte("4"))()
	if err != nil {
		panic(err)
	}
	Equal(t, string(val[0]), "5")
	val, err = c.KeyAddress([]byte("3"))()
	if err != nil {
		panic(err)
	}
	Equal(t, string(val[0]), "3")
}

func TestLargeAddress(t *testing.T) {
	c := RingFromList([]string{"b", "c", "a", "y"})
	val, err := c.KeyAddress([]byte("z"))()
	if err != nil {
		panic(err)
	}
	Equal(t, string(val[0]), "a")
}

func TestAdjacent(t *testing.T) {
	c := RingFromList([]string{
		"a",
		"b",
		"c",
		"d",
	})
	Equal(
		t,
		c.Adjacent(
			[]byte("a"),
			[]byte("b"),
		),
		true,
	)
	Equal(
		t,
		c.Adjacent(
			[]byte("a"),
			[]byte("c"),
		),
		false,
	)
	Equal(
		t,
		c.Adjacent(
			[]byte("d"),
			[]byte("a"),
		),
		true,
	)
}

func EqualRings(t *testing.T, c1 *Ring, c2 *Ring) {
	for current1, current2 := c1.next, c2.next; bytes.Compare(current1.address, nil) != 0; current1, current2 = current1.next, current2.next {
		if bytes.Compare(current1.address, current2.address) != 0 {
			t.Errorf("Expected %s, got %s", string(current1.address), string(current2.address))
		}
	}
}

func TestEqualRings(t *testing.T) {
	c := RingFromList([]string{
		"a",
		"b",
		"c",
		"d",
	})
	a := RingFromList([]string{
		"c",
		"d",
		"b",
		"a",
	})
	EqualRings(t, a, c)
	EqualRings(t, c, a)
}

func TestRemove(t *testing.T) {
	circle := RingFromList([]string{
		"a",
		"b",
		"c",
		"d",
	})
	step_1 := RingFromList([]string{
		"a",
		"c",
		"d",
	})
	step_2 := RingFromList([]string{
		"a",
		"c",
	})
	step_3 := RingFromList([]string{
		"c",
	})

	circle.RemoveString("b")
	EqualRings(t, step_1, circle)

	circle.RemoveString("d")
	EqualRings(t, step_2, circle)

	circle.RemoveString("a")
	EqualRings(t, step_3, circle)
}
