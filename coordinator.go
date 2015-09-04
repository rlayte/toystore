package toystore

import (
	"log"

	"github.com/rlayte/toystore/data"
)

// isCoordinator returns true if the current node is the owner
// of the provided address. Otherwise it returns false.
func (t *Toystore) isCoordinator(address string) bool {
	return address == t.rpcAddress()
}

// CoordinateGet organizes the get request between the collaborating nodes.
// It sends get requests to all nodes in the key's preference list and keeps
// track of success/failures. If there are more successful reads than config.R
// it returns the value and true. Otherwise it returns the value and false.
func (t *Toystore) CoordinateGet(key string) (*data.Data, bool) {
	log.Printf("Coordinating GET request %s.", key)

	var value *data.Data
	var ok bool

	nodes := t.Ring.FindN(key, t.ReplicationLevel)
	reads := 0

	for _, address := range nodes {
		if address != t.rpcAddress() {
			value, ok = t.client.Get(address, key)

			if ok {
				reads++
			}
		} else {
			log.Printf("Coordinator retrieving %s", key)
			value, ok = t.Data.Get(key)

			if ok {
				reads++
			}
		}
	}

	if reads < t.R {
		log.Printf("Reads too few %s for %s", reads, key)
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
	log.Printf("Coordinating PUT request %v", value)

	nodes := t.Ring.FindN(key, t.ReplicationLevel)
	writes := 0

	for address, hint := range nodes {
		if address != t.rpcAddress() {
			var ok bool

			if hint != address {
				ok = t.client.HintPut(address, string(hint), value)
			} else {
				ok = t.client.Put(address, value)
			}

			if ok {
				writes++
			}
		} else {
			log.Printf("Coordinator saving %s", value)
			ok := t.Data.Put(value)

			if ok {
				writes++
			}
		}
	}

	if writes < t.W {
		log.Printf("Writes too few %s for %s", writes, key)
	}

	return writes >= t.W
}
