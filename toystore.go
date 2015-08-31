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
	Port    int
	RPCPort int
	Host    string
	Data    Store
	Ring    *Ring
	Members Members

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
func (t *Toystore) Get(key string) (value string, ok bool) {
	lookup := t.Ring.KeyAddress([]byte(key))
	address, _ := lookup()

	log.Println("Get request", key, string(address))

	if t.isCoordinator(address) {
		value, ok = t.CoordinateGet(key)
	} else {
		value, ok = t.client.CoordinateGet(string(address), key)
	}

	log.Println("Get request", key, ok, t.isCoordinator(address), string(address), t.Host)
	return
}

func (t *Toystore) Put(key string, value string) (ok bool) {
	lookup := t.Ring.KeyAddress([]byte(key))
	address, _ := lookup()

	log.Println(t.Host, "Current members", t.Ring.AddressList(), string(address))

	if t.isCoordinator(address) {
		ok = t.CoordinatePut(key, value)
	} else {
		ok = t.client.CoordinatePut(string(address), key, value)
	}
	return
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
			t.client.CoordinatePut(string(address), key, val)
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

func New(config Config, seedMeta interface{}) *Toystore {
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

	ReplicationDepth = t.ReplicationLevel
	t.Ring.AddString(t.rpcAddress())
	NewRpcHandler(t)

	return t
}
