package toystore

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
)

var Zero []byte = make([]byte, 256)

type Status int

const (
	Alive Status = iota
	Dead
)

var Hash func([]byte) []byte = func(bytes []byte) []byte {
	hash := sha256.New()
	hash.Write(bytes)
	return hash.Sum(nil)
}

type Ring struct {
	address []byte
	hash    []byte
	time    *sync.Mutex
	status  Status
	next    *Ring
}

func (r *Ring) String() string {
	var buffer bytes.Buffer
	for current, first := r, true; len(current.hash) != 0 ||
		first; current, first = current.next, false {
		if !first {
			buffer.WriteString(" -> ")
		}
		buffer.WriteString(string(current.address))
		buffer.WriteString("/")
		// buffer.WriteString(string(current.hash))
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
	ring.status = Dead
	ring.time = &sync.Mutex{}
	ring.next = ring
	return ring
}

func NewRing(address []byte) *Ring {
	ring := new(Ring)
	ring.status = Alive
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
	r.time.Lock()
	for current = r.next; bytes.Compare(current.address, address) != 0; current = current.next {
		if string(current.hash) == "" {
			return errors.New(fmt.Sprintf("No such node in circle: %s\n", address))
		}
	}
	current.status = Dead
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

func (r *Ring) Address(key []byte) *Ring {
	current := r

	for bytes.Compare(current.hash, key) < 1 {
		current = current.next
	}

	return current.next
}

func (r *Ring) KeyAddress(key []byte) func() ([]byte, []byte, error) {
	hashed := Hash(key)

	r.time.Lock()
	current := r.find(hashed) // there should be no chance of this being dead.

	if bytes.Compare(current.hash, nil) == 0 {
		current = current.next
	}

	i := 0
	r.time.Unlock()
	return func() ([]byte, []byte, error) {
		r.time.Lock()
		defer r.time.Unlock()
		output := current.address
		i++

		if i > ReplicationDepth {
			return []byte{}, nil, errors.New("No more replications.")
		}

		current = current.next
		for current.status == Dead {
			current = current.next
		}

		return output, nil, nil
	}
}

func (r *Ring) find(address []byte) *Ring {
	var current *Ring
	hash := Hash(address)
	for current = r.next; bytes.Compare(current.hash, nil) != 0 &&
		bytes.Compare(current.hash, hash) == -1; current = current.next {
	}
	return current
}

func (r *Ring) Adjacent(first []byte, second []byte) bool {
	r.time.Lock()
	next := r.find(first).next
	for next.status == Dead { // guarentees not the head node.
		next = next.next // only hangs if everything is dead.
	}
	r.time.Unlock()
	return bytes.Compare(next.address, second) == 0
}
