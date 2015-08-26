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
	a := NewCircleHead()
	b := a.Add(NewCircleString("b"))
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
	c := CircleFromList([]string{"1", "3", "5"})
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
	c := CircleFromList([]string{"b", "c", "a", "y"})
	val, err := c.KeyAddress([]byte("z"))()
	if err != nil {
		panic(err)
	}
	Equal(t, string(val[0]), "a")
}

func TestAdjacent(t *testing.T) {
	c := CircleFromList([]string{
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

func EqualCircles(t *testing.T, c1 *Circle, c2 *Circle) {
	for current1, current2 := c1.next, c2.next; bytes.Compare(current1.address, nil) != 0; current1, current2 = current1.next, current2.next {
		if bytes.Compare(current1.address, current2.address) != 0 {
			t.Errorf("Expected %s, got %s", string(current1.address), string(current2.address))
		}
	}
}

func TestEqualCircles(t *testing.T) {
	c := CircleFromList([]string{
		"a",
		"b",
		"c",
		"d",
	})
	a := CircleFromList([]string{
		"c",
		"d",
		"b",
		"a",
	})
	EqualCircles(t, a, c)
	EqualCircles(t, c, a)
}

func TestRemove(t *testing.T) {
	circle := CircleFromList([]string{
		"a",
		"b",
		"c",
		"d",
	})
	step_1 := CircleFromList([]string{
		"a",
		"c",
		"d",
	})
	step_2 := CircleFromList([]string{
		"a",
		"c",
	})
	step_3 := CircleFromList([]string{
		"c",
	})

	circle.RemoveString("b")
	EqualCircles(t, step_1, circle)

	circle.RemoveString("d")
	EqualCircles(t, step_2, circle)

	circle.RemoveString("a")
	EqualCircles(t, step_3, circle)
}
