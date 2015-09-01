package memory

import "github.com/rlayte/toystore/data"

type MemoryStore struct {
	data map[string]*data.Data
}

func (m MemoryStore) Get(key string) (*data.Data, bool) {
	value, ok := m.data[key]
	return value, ok
}

func (m MemoryStore) Put(d *data.Data) bool {
	m.data[d.Key] = d
	return true
}

func New() *MemoryStore {
	return &MemoryStore{map[string]*data.Data{}}
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
