package main

import (
	"os"
	"strconv"

	"github.com/rlayte/toystore"
)

type MemoryStore struct {
	data map[string]string
}

func (m MemoryStore) Get(key string) (string, bool) {
	value, ok := m.data[key]
	return value, ok
}

func (m MemoryStore) Put(key string, value string) {
	m.data[key] = value
}

func main() {
	var seed string
	port, err := strconv.Atoi(os.Args[1])

	if port != 3000 {
		seed = ":3010"
	}

	if err != nil {
		panic(err)
	}

	t := toystore.New(port, &MemoryStore{map[string]string{}}, seed)
	t.Serve()
}
