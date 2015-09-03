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
	// Number of nodes to use as replicas
	ReplicationDepth int

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
	defer r.time.Unlock()
	for current, first := r, true; len(current.hash) != 0 ||
		first; current, first = current.next, false {

		output = append(output, string(current.address))
	}
	return output
}

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
	defer r.time.Unlock()
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
	defer r.time.Unlock()
	for current = r.next; bytes.Compare(current.address, address) != 0; current = current.next {
		if string(current.hash) == "" {
			return errors.New(fmt.Sprintf("No such node in circle: %s\n", address))
		}
	}
	current.status = Dead
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
	r.time.Lock()
	defer r.time.Unlock()

	hashed := Hash(key)
	current := r.find(hashed) // there should be no chance of this being dead.

	if bytes.Compare(current.hash, nil) == 0 {
		current = current.next
	}

	i := 0
	return func() ([]byte, []byte, error) {
		r.time.Lock()
		defer r.time.Unlock()

		i++
		if i > r.ReplicationDepth {
			return nil, nil, errors.New("No more replications.")
		}

		if bytes.Compare(current.address, nil) == 0 {
			current = current.next
		}

		// Find the hint if there is one.
		var hint *Ring = current
		for hint.status == Dead {
			hint = hint.next
		}

		// if we want to hint:
		if current.status == Dead {
			return hint.address, current.address, nil
		}

		return current.address, nil, nil
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
	defer r.time.Unlock()
	next := r.find(first).next
	for next.status == Dead { // guarentees not the head node.
		next = next.next // only hangs if everything is dead.
	}
	return bytes.Compare(next.address, second) == 0
}
