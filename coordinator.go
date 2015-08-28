package toystore

import "log"

func (t *Toystore) isCoordinator(address []byte) bool {
	return string(address) == t.rpcAddress()
}

func (t *Toystore) CoordinateGet(key string) (string, bool) {
	log.Printf("%s coordinating GET request %s.", t.Address(), key)

	var value string
	var ok bool

	lookup := t.Ring.KeyAddress([]byte(key))
	reads := 0

	for address, err := lookup(); err == nil; address, err = lookup() {
		if string(address) != t.rpcAddress() {
			log.Printf("%s sending GET request to %s.", t.Address(), address)
			value, ok = GetCall(string(address), key)

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

	return value, ok && reads >= t.R
}

func (t *Toystore) CoordinatePut(key string, value string) bool {
	log.Printf("%s coordinating PUT request %s/%s.", t.Address(), key, value)

	lookup := t.Ring.KeyAddress([]byte(key))
	writes := 0

	for address, err := lookup(); err == nil; address, err = lookup() {
		if string(address) != t.rpcAddress() {
			log.Printf("%s sending replication request to %s.", t.Address(), address)
			ok := PutCall(string(address), key, value)

			if ok {
				writes++
			}
		} else {
			log.Printf("Coordinator %s saving %s/%s.", t.Address(), key, value)
			ok := t.Data.Put(key, value)

			if ok {
				writes++
			}
		}
	}

	return writes >= t.W
}
