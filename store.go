package toystore

type Store interface {
	Get(string) (string, bool)
	Put(string, string) bool
	Keys() []string
}
