package toystore

import (
	"fmt"
	"log"

	"github.com/rlayte/toystore/data"
)

func (t *Toystore) isCoordinator(address []byte) bool {
	return string(address) == t.rpcAddress()
}

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

	//
	// Do we just take the last value?
	// There should be data-version resolution here!
	//

	value_string := fmt.Sprint(value)

	log.Println("Coordinate get complete", value_string, ok, reads)

	return value, ok && reads >= t.R
}

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
