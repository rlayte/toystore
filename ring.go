package toystore

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
)

var Zero []byte = make([]byte, 256)

var Hash func([]byte) []byte = func(bytes []byte) []byte {
	hash := sha256.New()
	hash.Write(bytes)
	return hash.Sum(nil)
}

type Circle struct {
	address []byte
	hash    []byte
	next    *Circle
}

func (c *Circle) String() string {
	var buffer bytes.Buffer
	for current, first := c, true; len(current.hash) != 0 ||
		first; current, first = current.next, false {
		if !first {
			buffer.WriteString(" -> ")
		}
		buffer.Write(current.address)
		buffer.WriteString("/")
		buffer.Write(current.hash)
	}
	return buffer.String()
}

func (c *Circle) AddressList() []string {
	output := make([]string, 0)
	for current, first := c, true; len(current.hash) != 0 ||
		first; current, first = current.next, false {

		output = append(output, string(current.address))
	}
	return output
}

var (
	ReplicationDepth int = 1
)

func NewCircleHead() *Circle {
	circle := new(Circle)
	circle.hash = []byte{} // empty is head.
	circle.next = circle
	return circle
}

func NewCircle(address []byte) *Circle {
	circle := new(Circle)
	circle.address = address
	circle.hash = Hash(address)
	return circle
}

func NewCircleString(address string) *Circle {
	return NewCircle([]byte(address))
}

func (c *Circle) Add(incoming *Circle) *Circle {
	var current *Circle
	for current = c; bytes.Compare(current.next.hash, incoming.hash) == -1; current = current.next {
		if bytes.Compare(current.next.hash, incoming.hash) == 0 {
			return nil
		}
		if bytes.Compare(current.next.hash, nil) == 0 {
			break
		}
	}
	incoming.next = current.next
	current.next = incoming
	return incoming
}

func (c *Circle) AddString(address string) *Circle {
	return c.Add(NewCircleString(address))
}

func (c *Circle) RemoveString(address string) error {
	return c.Remove([]byte(address))
}

func (c *Circle) Remove(address []byte) error {
	var current *Circle
	var last *Circle
	for current, last = c.next, c; bytes.Compare(current.address, address) != 0; current, last = current.next, current {
		if string(current.hash) == "" {
			return errors.New(fmt.Sprintf("No such node in circle: %s\n", address))
		}
	}
	last.next = current.next
	return nil
}

func CircleFromList(strs []string) *Circle {
	circle := NewCircleHead()
	for _, str := range strs {
		circle.AddString(str)
	}
	return circle
}

func (c *Circle) KeyAddress(key []byte) func() ([]byte, error) {
	hashed := Hash(key)

	current := c.find(hashed)

	if bytes.Compare(current.hash, nil) == 0 {
		current = current.next
	}

	i := 0
	return func() ([]byte, error) {
		output := current.address
		i++

		if i > ReplicationDepth {
			return []byte{}, errors.New("No more replications.")
		}

		current = current.next
		if bytes.Compare(current.hash, nil) == 0 {
			current = current.next
		}

		return output, nil
	}
}

func (c *Circle) find(address []byte) *Circle {
	var current *Circle
	for current = c.next; bytes.Compare(current.hash, nil) != 0 &&
		bytes.Compare(current.address, address) == -1; current = current.next {
	}
	return current
}

func (c *Circle) Adjacent(first []byte, second []byte) bool {
	next := c.find(first).next
	if next.address == nil {
		next = next.next
	}
	return bytes.Compare(next.address, second) == 0
}
