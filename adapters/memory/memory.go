package memory

type MemoryStore struct {
	data map[string]string
}

func (m MemoryStore) Get(key string) (string, bool) {
	value, ok := m.data[key]
	return value, ok
}

func (m MemoryStore) Put(key string, value string) bool {
	m.data[key] = value
	return true
}

func New() *MemoryStore {
	return &MemoryStore{map[string]string{}}
}
