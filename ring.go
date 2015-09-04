package toystore

import (
	"bytes"
	"container/list"
	"crypto/sha256"
	"log"
	"strings"
	"sync"
)

// Hash is the hashing function used to determine a nodes position in the
// ring and to find the appropriate node for a given key.
var Hash func([]byte) []byte = func(bytes []byte) []byte {
	hash := sha256.New()
	hash.Write(bytes)
	return hash.Sum(nil)
}

// lessThan returns true is the hashed value of a is less than the hashed
// value of b. Otherwise it returns false.
func lessThan(a *list.Element, b string) bool {
	return bytes.Compare([]byte(a.Value.(string)), Hash([]byte(b))) < 0
}

// HashRing maintains a list of members and their position in the cluster
// based on a consistent hashing function.
// If a node fails it remains in the list but is marked as failed.
type HashRing struct {
	list   *list.List
	failed map[string]bool
	lock   *sync.Mutex
}

// findElement iterates over the node until it finds a node greater
// than the key.
// If a node isn't found it returns the head of the list.
func (h *HashRing) findElement(key string) *list.Element {
	current := h.list.Front()

	for lessThan(current, key) {
		current = current.Next()

		if current == nil {
			break
		}
	}

	return current
}

// String returns a comma separated list of addresses.
func (h *HashRing) String() string {
	current := h.list.Front()
	addresses := []string{}

	for current != nil {
		addresses = append(addresses, current.Value.(string))
		current = current.Next()
	}

	return strings.Join(addresses, ", ")
}

// Add finds the first node that is higher than the address and inserts
// a new node before it.
func (h *HashRing) Add(address string) {
	if h.list.Len() == 0 {
		h.list.PushBack(address)
	} else {
		target := h.findElement(address)

		if target != nil {
			h.list.InsertBefore(address, target)
		} else {
			h.list.PushBack(address)
		}
	}

	log.Printf("Addded %s. Size: %d", address, h.list.Len())
}

// Find returns the node that owns the range the key falls within.
// TODO: Should this return hinted addresses if the node is dead?
func (h *HashRing) Find(key string) string {
	log.Printf("Finding address for %s in %s", key, h)
	element := h.findElement(key)

	if element != nil {
		return element.Value.(string)
	} else {
		return ""
	}
}

// FindN returns n alive nodes starting with the closest to the provided key.
// If a node is dead the next alive node will be returned in its place with a
// hint to the real address.
// Returns a map where keys are addresses and values are hints. If the key and
// value are the same then the node is alive.
func (h *HashRing) FindN(key string, n int) map[string]string {
	log.Printf("Finding addresses for %s in %s", key, h)
	target := h.findElement(key)
	ret := map[string]string{}

	for len(ret) < n {
		address := target.Value.(string)
		hint := address

		for h.failed[address] {
			address = target.Next().Value.(string)
		}

		ret[address] = hint
		target = target.Next()

		if target == nil {
			target = h.list.Front()
		}
	}

	return ret
}

// Adjacent returns true if the addresses are next to each other in the ring.
func (h *HashRing) Adjacent(a, b string) bool {
	nodeA := h.findElement(a)
	nodeB := h.findElement(b)

	if nodeA != nil {
		next := nodeA.Next()

		if next == nil {
			next = h.list.Front()
		}

		return next == nodeB
	} else {
		return false
	}
}

// Fail marks member as failed, but doesn't remove it from the ring.
func (h *HashRing) Fail(member string) {
	h.failed[member] = true
}

// Revive removes the member from the failed list.
func (h *HashRing) Revive(member string) {
	delete(h.failed, member)
}

func NewHashRing() *HashRing {
	return &HashRing{
		list:   list.New(),
		failed: map[string]bool{},
		lock:   &sync.Mutex{},
	}
}
