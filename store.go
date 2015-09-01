package toystore

import "github.com/rlayte/toystore/data"

type Store interface {
	Get(string) (*data.Data, bool)
	Put(*data.Data) bool
	Keys() []string
}
