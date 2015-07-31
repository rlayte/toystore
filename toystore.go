package toystore

import (
	"encoding/gob"
	"fmt"
	"log"
	"time"

	"github.com/charlesetc/circle"
	"github.com/charlesetc/dive"
)

type Toystore struct {
	// Config
	ReplicationLevel int
	W                int
	R                int

	// Internal use
	dive *dive.Node
	Port int
	Data Store
	Ring *circle.Circle
}

type ToystoreMetaData struct {
	Address    string
	RPCAddress string
}

type Store interface {
	Get(string) (string, bool)
	Put(string, string) bool
	Keys() []string
}

func (t *Toystore) UpdateMembers() {
	addresses := []string{t.rpcAddress()}

	for _, member := range t.dive.Members {
		if member.MetaData != nil {
			metaData := member.MetaData.(ToystoreMetaData)
			addresses = append(addresses, metaData.RPCAddress)
		}
	}

	t.Ring = circle.CircleFromList(addresses)
}

func (t *Toystore) Address() string {
	return fmt.Sprintf(":%d", t.Port)
}

func (t *Toystore) rpcAddress() string {
	return fmt.Sprintf(":%d", t.Port+20)
}

func RpcToAddress(rpc string) string {
	var port int
	fmt.Sscanf(rpc, ":%d", &port)
	return fmt.Sprintf(":%d", port-20)
}

func (t *Toystore) CoordinateGet(key string) (string, bool) {
	t.UpdateMembers()
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

func (t *Toystore) Get(key string) (string, bool) {
	lookup := t.Ring.KeyAddress([]byte(key))
	address, _ := lookup()

	var value string
	var ok bool

	if string(address) != t.rpcAddress() {
		log.Printf("%s forwarding GET request to %s. %s", t.Address(), address, key)
		value, ok = CoordinateGet(string(address), key)
	} else {
		value, ok = t.CoordinateGet(key)
	}
	return value, ok
}

func (t *Toystore) CoordinatePut(key string, value string) bool {
	t.UpdateMembers()

	log.Printf("%s coordinating PUT request %s/%s.", t.Address(), key, value)

	lookup := t.Ring.KeyAddress([]byte(key))
	writes := 0

	for address, err := lookup(); err == nil; address, err = lookup() {
		if string(address) != t.rpcAddress() {
			log.Printf("%s sending replation request to %s.", t.Address(), address)
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

func (t *Toystore) Put(key string, value string) bool {
	lookup := t.Ring.KeyAddress([]byte(key))
	address, _ := lookup()

	var ok bool

	if string(address) != t.rpcAddress() {
		log.Printf("%s forwarding PUT request to coordinator %s.", t.Address(), address)
		ok = CoordinatePut(string(address), key, value)
	} else {
		ok = t.CoordinatePut(key, value)
	}
	return ok
}

func New(port int, store Store, seed string, seedMeta interface{}) *Toystore {
	t := &Toystore{
		ReplicationLevel: 3,
		W:                1,
		R:                1,
		Port:             port,
		Data:             store,
	}

	circle.ReplicationDepth = t.ReplicationLevel

	dive.PingInterval = time.Second
	n := dive.NewNode(port+10, &dive.BasicRecord{Address: seed, MetaData: seedMeta})
	n.MetaData = ToystoreMetaData{t.Address(), t.rpcAddress()}
	gob.RegisterName("ToystoreMetaData", n.MetaData)

	t.dive = n
	return t
}
