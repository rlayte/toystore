package toystore

import (
	"bytes"
	"container/list"
	"crypto/sha256"
	"sync"
)

var Hash func([]byte) []byte = func(bytes []byte) []byte {
	hash := sha256.New()
	hash.Write(bytes)
	return hash.Sum(nil)
}

type HashRing struct {
	list   *list.List
	failed map[string]bool
	lock   *sync.Mutex
}

func lessThan(a *list.Element, b string) bool {
	return bytes.Compare([]byte(a.Value.(string)), Hash([]byte(b))) < 1
}

func (h *HashRing) findElement(key string) *list.Element {
	current := h.list.Front()

	for lessThan(current, key) {
		current = current.Next()

		if current == nil {
			current = h.list.Front()
			break
		}
	}

	return current
}

func (h *HashRing) Add(address string) {
	h.lock.Lock()
	defer h.lock.Unlock()

	if h.list.Len() == 0 {
		h.list.PushBack(address)
		return
	} else {
		target := h.findElement(address)
		h.list.InsertBefore(address, target)
	}
}

func (h *HashRing) Find(key string) string {
	return h.findElement(key).Value.(string)
}

func (h *HashRing) FindN(key string, n int) map[string]string {
	h.lock.Lock()
	defer h.lock.Unlock()
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
	}

	return ret
}

func (h *HashRing) Adjacent(a, b string) bool {
	h.lock.Lock()
	defer h.lock.Unlock()
	return h.findElement(a).Next() == h.findElement(b)
}

func (h *HashRing) Fail(member string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.failed[member] = true
}

func (h *HashRing) Revive(member string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	delete(h.failed, member)
}

func NewHashRing() *HashRing {
	return &HashRing{
		list:   list.New(),
		failed: map[string]bool{},
		lock:   &sync.Mutex{},
	}
}
