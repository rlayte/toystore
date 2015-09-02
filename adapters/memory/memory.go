package memory

import (
	"sync"

	"github.com/rlayte/toystore/data"
)

type MemoryStore struct {
	lock *sync.Mutex
	data map[string]*data.Data
}

func (m MemoryStore) Get(key string) (*data.Data, bool) {
	m.lock.Lock()
	value, ok := m.data[key]
	m.lock.Unlock()
	return value, ok
}

func (m MemoryStore) Put(d *data.Data) bool {
	m.lock.Lock()
	m.data[d.Key] = d
	m.lock.Unlock()
	return true
}

func New() *MemoryStore {
	return &MemoryStore{&sync.Mutex{}, map[string]*data.Data{}}
}

func (m MemoryStore) Keys() []string {
	out := make([]string, len(m.data))
	i := 0
	for key := range m.data {
		out[i] = key
		i++
	}
	return out
}
