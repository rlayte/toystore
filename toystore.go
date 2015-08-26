package toystore

import (
	"encoding/gob"
	"fmt"
	"log"
	"time"

	"github.com/Charlesetc/circle"
	"github.com/charlesetc/dive"
)

type Toystore struct {
	// Config
	ReplicationLevel int
	W                int
	R                int

	// Internal use
	dive            *dive.Node
	Port            int
	Data            Store
	Ring            *Circle
	request_address chan []byte
	receive_address chan func() ([]byte, error) // will eventually not be bool anymore.
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

	t.Ring = CircleFromList(addresses)
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

func (t *Toystore) isCoordinator(address []byte) bool {
	return string(address) == t.rpcAddress()
}

func (t *Toystore) CoordinateGet(key string) (string, bool) {

	log.Printf("%s coordinating GET request %s.", t.Address(), key)

	var value string
	var ok bool

	lookup := t.KeyAddress([]byte(key))
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

// An exposed endpoint to the client.
// Should function by directing each get or put
// to the proper machine.
func (t *Toystore) Get(key string) (value string, ok bool) {
	// t.UpdateMembers()

	lookup := t.KeyAddress([]byte(key))
	address, _ := lookup()

	// if this is the right node...
	if t.isCoordinator(address) {
		// take care of the get myself
		value, ok = t.CoordinateGet(key)
	} else {
		// send it off to the right one.
		value, ok = CoordinateGetCall(string(address), key)
	}
	return
}

func (t *Toystore) CoordinatePut(key string, value string) bool {

	log.Printf("%s coordinating PUT request %s/%s.", t.Address(), key, value)

	lookup := t.KeyAddress([]byte(key))
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

func (t *Toystore) Put(key string, value string) (ok bool) {
	lookup := t.KeyAddress([]byte(key))
	address, _ := lookup()

	if t.isCoordinator(address) {
		ok = t.CoordinatePut(key, value)
	} else {
		ok = CoordinatePutCall(string(address), key, value)
	}
	return
}

func (t *Toystore) KeyAddress(key []byte) func() ([]byte, error) {
	t.request_address <- key
	f := <-t.receive_address
	return f
}

func (t *Toystore) Adjacent(address string) bool {
	return t.Ring.Adjacent([]byte(t.rpcAddress()), []byte(address))
}

func (t *Toystore) Transfer(address string) {
	keys := t.Data.Keys()
	for _, key := range keys {
		val, ok := t.Data.Get(key)
		if !ok {
			// Should not happen.
			panic("I was told this key existed but it doesn't...")
		}
		log.Printf("Forward %s/%s\n", key, string(val))
		// Checks to see if it's my key and if it's not, it forwards
		// the put call.
		lookup := t.Ring.KeyAddress([]byte(key))
		address, _ := lookup()

		if !t.isCoordinator(address) {
			CoordinatePutCall(string(address), key, val)
		}
	}
}

func (t *Toystore) handleJoin(address string) {
	log.Printf("Toystore joined: %s\n", address)
	t.Ring.AddString(address)

	if t.Adjacent(address) {
		log.Println("Adjacent.")
		t.Transfer(address)
	}
}

func (t *Toystore) handleFail(address string) {
	log.Printf("Toystore left: %s\n", address)
	if address != t.rpcAddress() {
		t.Ring.RemoveString(address) // this is causing a problem
	}
}

func (t *Toystore) serveAsync() {
	for {
		select {
		case event := <-t.dive.Events:
			switch event.Kind {
			case dive.Join:
				address := event.Data.(ToystoreMetaData).RPCAddress // might not be rpc..
				t.handleJoin(address)
			case dive.Fail:
				address := event.Data.(ToystoreMetaData).RPCAddress
				t.handleFail(address)
			}
		case key := <-t.request_address:
			log.Println("request_address")
			t.receive_address <- t.Ring.KeyAddress(key)
		}
	}
}

func New(port int, store Store, seed string, seedMeta interface{}) *Toystore {
	t := &Toystore{
		ReplicationLevel: 3,
		W:                1,
		R:                1,
		Port:             port,
		Data:             store,
		request_address:  make(chan []byte),
		receive_address:  make(chan func() ([]byte, error)),
		Ring:             NewCircleHead(),
	}

	circle.ReplicationDepth = t.ReplicationLevel

	dive.PingInterval = time.Second
	n := dive.NewNode(
		"localhost",
		port+10,
		&dive.BasicRecord{Address: seed, MetaData: seedMeta},
		make(chan *dive.Event),
	)
	n.MetaData = ToystoreMetaData{t.Address(), t.rpcAddress()}
	gob.RegisterName("ToystoreMetaData", n.MetaData)

	t.dive = n

	// Add yourself to the ring
	t.Ring.AddString(t.rpcAddress())

	go t.serveAsync()

	go ServeRPC(t)

	return t
}
