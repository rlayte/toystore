package toystore

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
)

var Zero []byte = make([]byte, 256)

var Hash func([]byte) []byte = func(bytes []byte) []byte {
	hash := sha256.New()
	hash.Write(bytes)
	return hash.Sum(nil)
}

type Ring struct {
	address []byte
	hash    []byte
	time    *sync.Mutex
	next    *Ring
}

func (r *Ring) String() string {
	var buffer bytes.Buffer
	for current, first := r, true; len(current.hash) != 0 ||
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

func (r *Ring) AddressList() []string {
	output := make([]string, 0)
	r.time.Lock()
	for current, first := r, true; len(current.hash) != 0 ||
		first; current, first = current.next, false {

		output = append(output, string(current.address))
	}
	r.time.Unlock()
	return output
}

var (
	ReplicationDepth int = 1
)

func NewRingHead() *Ring {
	ring := new(Ring)
	ring.hash = []byte{} // empty is head.
	ring.time = &sync.Mutex{}
	ring.next = ring
	return ring
}

func NewRing(address []byte) *Ring {
	ring := new(Ring)
	ring.address = address
	ring.hash = Hash(address)
	return ring
}

func NewRingString(address string) *Ring {
	return NewRing([]byte(address))
}

func (r *Ring) Add(incoming *Ring) *Ring {
	var current *Ring
	r.time.Lock()
	for current = r; bytes.Compare(current.next.hash, incoming.hash) == -1; current = current.next {
		if bytes.Compare(current.next.hash, incoming.hash) == 0 {
			return nil
		}
		if bytes.Compare(current.next.hash, nil) == 0 {
			break
		}
	}
	incoming.next = current.next
	current.next = incoming
	r.time.Unlock()
	return incoming
}

func (r *Ring) AddString(address string) *Ring {
	return r.Add(NewRingString(address))
}

func (r *Ring) RemoveString(address string) error {
	return r.Remove([]byte(address))
}

func (r *Ring) Remove(address []byte) error {
	var current *Ring
	var last *Ring
	r.time.Lock()
	for current, last = r.next, r; bytes.Compare(current.address, address) != 0; current, last = current.next, current {
		if string(current.hash) == "" {
			return errors.New(fmt.Sprintf("No such node in circle: %s\n", address))
		}
	}
	last.next = current.next
	r.time.Unlock()
	return nil
}

func RingFromList(strs []string) *Ring {
	ring := NewRingHead()
	for _, str := range strs {
		ring.AddString(str)
	}
	return ring
}

func (r *Ring) KeyAddress(key []byte) func() ([]byte, error) {
	hashed := Hash(key)

	r.time.Lock()
	current := r.find(hashed)

	if bytes.Compare(current.hash, nil) == 0 {
		current = current.next
	}

	i := 0
	r.time.Unlock()
	return func() ([]byte, error) {
		r.time.Lock()
		output := current.address
		i++

		if i > ReplicationDepth {
			return []byte{}, errors.New("No more replications.")
		}

		current = current.next
		if bytes.Compare(current.hash, nil) == 0 {
			current = current.next
		}
		r.time.Unlock()
		return output, nil
	}
}

func (c *Ring) find(address []byte) *Ring {
	var current *Ring
	for current = c.next; bytes.Compare(current.hash, nil) != 0 &&
		bytes.Compare(current.address, address) == -1; current = current.next {
	}
	return current
}

func (c *Ring) Adjacent(first []byte, second []byte) bool {
	c.time.Lock()
	next := c.find(first).next
	if next.address == nil {
		next = next.next
	}
	c.time.Unlock()
	return bytes.Compare(next.address, second) == 0
}
