package toystore

import (
	"fmt"
	"log"
)

type Toystore struct {
	// Config
	ReplicationLevel int
	W                int
	R                int

	// Internal use
	Port           int
	RPCPort        int
	Host           string
	Data           Store
	Ring           *Ring
	Members        Members
	requestAddress chan []byte
	receiveAddress chan func() ([]byte, error)
}

type ToystoreMetaData struct {
	Address    string
	RPCAddress string
}

func (t *Toystore) Address() string {
	return fmt.Sprintf("%s:%d", t.Host, t.Port)
}

func (t *Toystore) rpcAddress() string {
	return fmt.Sprintf("%s:%d", t.Host, t.RPCPort)
}

// An exposed endpoint to the client.
// Should function by directing each get or put
// to the proper machine.
func (t *Toystore) Get(key string) (value string, ok bool) {
	lookup := t.KeyAddress([]byte(key))
	address, _ := lookup()

	if t.isCoordinator(address) {
		value, ok = t.CoordinateGet(key)
	} else {
		value, ok = CoordinateGetCall(string(address), key)
	}
	return
}

func (t *Toystore) Put(key string, value string) (ok bool) {
	log.Println("Putting", t.Members.Len())
	for _, member := range t.Members.Members() {
		log.Printf("%s has member %s", t.Address(), member.Name())
	}

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
	t.requestAddress <- key
	f := <-t.receiveAddress
	return f
}

func (t *Toystore) Transfer(address string) {
	keys := t.Data.Keys()
	for _, key := range keys {
		val, ok := t.Data.Get(key)
		if !ok {
			panic("I was told this key existed but it doesn't...")
		}
		log.Printf("Forward %s/%s\n", key, string(val))
		lookup := t.Ring.KeyAddress([]byte(key))
		address, _ := lookup()

		if !t.isCoordinator(address) {
			CoordinatePutCall(string(address), key, val)
		}
	}
}

func (t *Toystore) AddMember(member Member) {
	t.Ring.AddString(member.Name())
	localAddress := t.rpcAddress()
	adjacent := t.Ring.Adjacent([]byte(localAddress), member.Meta())

	if adjacent {
		log.Println("Adjacent.")
		t.Transfer(member.Address())
	}
}

func (t *Toystore) RemoveMember(member Member) {
	if member.Address() != t.rpcAddress() {
		t.Ring.RemoveString(member.Name()) // this is causing a problem
	}
}

func New(config Config, seedMeta interface{}) *Toystore {
	t := &Toystore{
		ReplicationLevel: config.ReplicationLevel,
		W:                config.W,
		R:                config.R,
		Host:             config.Host,
		Port:             config.ClientPort,
		RPCPort:          config.RPCPort,
		Data:             config.Store,
		requestAddress:   make(chan []byte),
		receiveAddress:   make(chan func() ([]byte, error)),
		Ring:             NewRingHead(),
	}

	t.Members = NewMemberlist(t, config.SeedAddress)

	ReplicationDepth = t.ReplicationLevel
	t.Ring.AddString(t.rpcAddress())
	go t.serveAsync()
	go ServeRPC(t)

	return t
}
