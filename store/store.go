package store

import "github.com/rlayte/toystore/data"

// Store should be implemented to persist data using a specific storage backend.
// See the adapters package for examples.
type Store interface {
	Get(string) (*data.Data, bool)
	Put(*data.Data) bool
	Keys() []string
}
