package toystore

import (
	"encoding/gob"
	"fmt"
	"log"

	"github.com/rlayte/toystore/data"
)

type Toystore struct {
	// Config
	ReplicationLevel int
	W                int
	R                int

	// Internal use
	Port    int
	RPCPort int
	Host    string
	Data    Store
	Ring    *Ring
	Members Members
	Hints   *HintedHandoff

	client PeerClient
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
func (t *Toystore) Get(key string) (interface{}, bool) {
	lookup := t.Ring.KeyAddress([]byte(key))
	address, _, _ := lookup()
	var data *data.Data
	var ok bool

	if t.isCoordinator(address) {
		data, ok = t.CoordinateGet(key)
	} else {
		data, ok = t.client.CoordinateGet(string(address), key)
	}

	if ok {
		return data.Value, ok
	} else {
		return nil, ok
	}
}

func (t *Toystore) Put(key string, value interface{}) (ok bool) {
	lookup := t.Ring.KeyAddress([]byte(key))
	address, _, _ := lookup()

	if t.isCoordinator(address) {
		ok = t.CoordinatePut(data.New(key, value))
	} else {
		ok = t.client.CoordinatePut(string(address), data.New(key, value))
	}
	return
}

// Just in case you don't want to deal with Interfaces:
func (t *Toystore) GetString(key string) (string, bool) {
	d, ok := t.Get(key)
	return d.(string), ok
}

func (t *Toystore) PutString(key string, value string) bool {
	return t.Put(key, value) // Just a wrapper, but it gives type checking.
}

func (t *Toystore) Merge(data *data.Data) bool {
	// Only updates the store if the new record is later
	// Assumes store implementations are thread safe

	current, _ := t.Data.Get(data.Key)

	if data.IsLater(current) {
		t.Data.Put(data)
		return true
	}

	return false
}

// TODO: should use Transfer RPC
func (t *Toystore) Transfer(address string) {
	keys := t.Data.Keys()
	for _, key := range keys {
		val, ok := t.Data.Get(key)
		if !ok {
			panic("I was told this key existed but it doesn't...")
		}
		log.Printf("Forward %s/%s\n", key, fmt.Sprint(val.Value))
		lookup := t.Ring.KeyAddress([]byte(key))
		address, _, _ := lookup()

		if !t.isCoordinator(address) {
			t.client.CoordinatePut(string(address), val)
		}
	}
}

func (t *Toystore) AddMember(member Member) {
	log.Printf("%s adding member %s", t.Host, member.Name())
	t.Ring.AddString(member.Address())
	localAddress := t.rpcAddress()
	adjacent := t.Ring.Adjacent([]byte(localAddress), member.Meta())

	if adjacent {
		t.Transfer(member.Address())
	}
}

func (t *Toystore) RemoveMember(member Member) {
	if member.Address() != t.rpcAddress() {
		log.Printf("%s removing member %s", t.Host, member.Name())
		t.Ring.RemoveString(member.Address()) // this is causing a problem
	}
}

func New(config Config) *Toystore {
	t := &Toystore{
		ReplicationLevel: config.ReplicationLevel,
		W:                config.W,
		R:                config.R,
		Host:             config.Host,
		Port:             config.ClientPort,
		RPCPort:          config.RPCPort,
		Ring:             NewRingHead(),
		Data:             config.Store,

		client: NewRpcClient(),
	}

	t.Members = NewMemberlist(t, config.SeedAddress)
	t.Hints = NewHintedHandoff(config, t.client)

	gob.Register(data.Data{})
	ReplicationDepth = t.ReplicationLevel
	t.Ring.AddString(t.rpcAddress())
	NewRpcHandler(t)

	return t
}
