package toystore

import (
	"encoding/gob"
	"fmt"
	"log"
	"time"

	"github.com/charlesetc/dive"
)

type Toystore struct {
	// Config
	ReplicationLevel int
	W                int
	R                int

	// Internal use
	dive           *dive.Node
	Port           int
	RPCPort        int
	GossipPort     int
	Host           string
	Data           Store
	Ring           *Ring
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

func RpcToAddress(rpc string) string {
	var port int
	fmt.Sscanf(rpc, ":%d", &port)
	return fmt.Sprintf(":%d", port-20)
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

func New(config Config, seedMeta interface{}) *Toystore {
	t := &Toystore{
		ReplicationLevel: config.ReplicationLevel,
		W:                config.W,
		R:                config.R,
		Host:             config.Host,
		Port:             config.ClientPort,
		RPCPort:          config.RPCPort,
		GossipPort:       config.GossipPort,
		Data:             config.Store,
		requestAddress:   make(chan []byte),
		receiveAddress:   make(chan func() ([]byte, error)),
		Ring:             NewRingHead(),
	}

	ReplicationDepth = t.ReplicationLevel
	dive.PingInterval = time.Second

	seed := &dive.BasicRecord{Address: config.SeedAddress, MetaData: seedMeta}
	n := dive.NewNode(config.Host, config.GossipPort, seed)
	n.MetaData = ToystoreMetaData{t.Address(), t.rpcAddress()}
	gob.RegisterName("ToystoreMetaData", n.MetaData)

	t.dive = n

	t.Ring.AddString(t.rpcAddress())
	go t.serveAsync()
	go ServeRPC(t)

	return t
}
