package toystore

import (
	"fmt"
	"log"

	"github.com/rlayte/toystore/data"
)

// isCoordinator returns true if the current node is the owner
// of the provided address. Otherwise it returns false.
func (t *Toystore) isCoordinator(address []byte) bool {
	return string(address) == t.rpcAddress()
}

// CoordinateGet organizes the get request between the collaborating nodes.
// It sends get requests to all nodes in the key's preference list and keeps
// track of success/failures. If there are more successful reads than config.R
// it returns the value and true. Otherwise it returns the value and false.
func (t *Toystore) CoordinateGet(key string) (*data.Data, bool) {
	log.Printf("%s coordinating GET request %s.", t.Address(), key)

	var value *data.Data
	var ok bool

	lookup := t.Ring.KeyAddress([]byte(key))
	reads := 0

	for address, _, err := lookup(); err == nil; address, _, err = lookup() {
		if string(address) != t.rpcAddress() {
			log.Printf("%s sending GET request to %s.", t.Address(), address)
			value, ok = t.client.Get(string(address), key)

			if ok {
				reads++
			}
		} else {
			log.Printf("Coordinator %s retrieving %s.", t.Address(), key)
			value, ok = t.Data.Get(key)

			if ok {
				reads++
			}
		}
	}

	// TODO: should use data versioning
	return value, ok && reads >= t.R
}

// CoordinatePut organizes the put request between the collaborating nodes.
// It sends put requests to all nodes in the key's preference list and keeps
// track of success/failures. If there are more successful writes than config.W
// it returns true. Otherwise it returns false.
//
// If any nodes in the key's preference list are dead it will attempt to put
// the value on other nodes with a hint to its correct location.
func (t *Toystore) CoordinatePut(value *data.Data) bool {
	key := value.Key
	value_string := fmt.Sprint(value.Value)
	log.Printf("%s coordinating PUT request %s/%s.", t.Address(), key, value_string)

	lookup := t.Ring.KeyAddress([]byte(key))
	writes := 0

	for address, hint, err := lookup(); err == nil; address, hint, err = lookup() {
		if string(address) != t.rpcAddress() {
			log.Printf("%s sending PUT request to %s.", t.Address(), address)
			var ok bool

			if hint != nil {
				ok = t.client.HintPut(string(address), string(hint), value)
			} else {
				ok = t.client.Put(string(address), value)
			}

			if ok {
				writes++
			}
		} else {
			log.Printf("Coordinator %s saving %s/%s.", t.Address(), key, value_string)
			ok := t.Data.Put(value)

			if ok {
				writes++
			}
		}
	}

	return writes >= t.W
}
