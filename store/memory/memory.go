// Package memory is an example Store implementation that saves values to
// local memory.
package memory

import (
	"sync"

	"github.com/rlayte/toystore/data"
)

type MemoryStore struct {
	lock *sync.Mutex
	data map[string]*data.Data
}

// Get returns a Data value and existence bool for the given key.
// Thread safe.
func (m MemoryStore) Get(key string) (*data.Data, bool) {
	m.lock.Lock()
	value, ok := m.data[key]
	m.lock.Unlock()
	return value, ok
}

// Put adds a new Data value and returns a success status bool.
// Thread safe.
func (m MemoryStore) Put(d *data.Data) bool {
	m.lock.Lock()
	m.data[d.Key] = d
	m.lock.Unlock()
	return true
}

// Keys returns a list of all keys added to the store.
// Thread safe.
func (m MemoryStore) Keys() []string {
	out := make([]string, len(m.data))
	i := 0

	m.lock.Lock()
	for key := range m.data {
		out[i] = key
		i++
	}
	m.lock.Unlock()

	return out
}

// New returns a new MemoryStore instance, creating the required
// data structure and lock.
func New() *MemoryStore {
	return &MemoryStore{&sync.Mutex{}, map[string]*data.Data{}}
}
