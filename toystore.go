// Package toystore is a simple implementation of a Dynamo database -
// http://www.allthingsdistributed.com/files/amazon-dynamo-sosp2007.pdf
//
// For more information see the project's readme -
// https://github.com/rlayte/toystore
package toystore

import (
	"fmt"
	"log"

	"github.com/rlayte/toystore/data"
)

type Toystore struct {
	// Number of nodes to replicate each data item.
	ReplicationLevel int

	// Number of successful writes required.
	W int

	// Number of successful reads required.
	R int

	// Port number to serve RPC requests between nodes.
	RPCPort int

	// Host address to bind to.
	Host string

	// Concrete Store implementation to persist data.
	Data Store

	// Hash ring for nodes in the cluster.
	Ring *Ring

	// Concrete Members implementation to represent the current cluster state.
	Members Members

	// Store of hinted data meant for other nodes.
	Hints *HintedHandoff

	// Concrete PeerClient implementation to make calls to other nodes.
	client PeerClient
}

// rpcAddress returns a string for the RPC address.
func (t *Toystore) rpcAddress() string {
	return fmt.Sprintf("%s:%d", t.Host, t.RPCPort)
}

// Get finds the key on the correct node in the cluster and returns
// the value and an existence bool.
// If the key is on the current node then it coordinates the operation.
// Otherwise it sends the coordination request to the correct node.
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

// Put finds the key on the correct node in the cluster, sets
// the value and returns a status bool.
// If the key is owned by current node then it coordinates the operation.
// Otherwise it sends the coordination request to the correct node.
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

// GetString returns a string of the value for the specified key/value pair.
func (t *Toystore) GetString(key string) (string, bool) {
	d, ok := t.Get(key)
	return d.(string), ok
}

// Merge updates the data object only if its Timestamp is later than the
// current value.
// If the key doesn't exist it adds it.
// Requires Store implementation to be thread safe.
func (t *Toystore) Merge(data *data.Data) bool {
	current, ok := t.Data.Get(data.Key)

	if !ok || data.IsLater(current) {
		t.Data.Put(data)
		return true
	}

	return false
}

// Transfer sends a list of keys to another node in the cluster.
func (t *Toystore) Transfer(address string) {
	// TODO: should use Transfer RPC
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

// AddMember adds a new node to the hash ring.
// If the new node is adjacent to the current node then it transfers
// any keys in its range that should be owned by the new node.
func (t *Toystore) AddMember(member Member) {
	log.Printf("Adding member %s", member.Name())
	t.Ring.AddString(member.Address())
	localAddress := t.rpcAddress()
	adjacent := t.Ring.Adjacent([]byte(localAddress), member.Meta())

	if adjacent {
		t.Transfer(member.Address())
	}
}

// RemoveMember removes a member from the hash ring.
func (t *Toystore) RemoveMember(member Member) {
	if member.Address() != t.rpcAddress() {
		log.Printf("Removing member %s", member.Name())
		t.Ring.RemoveString(member.Address())
	}
}

// New creates a new Toystore instance using the config variables.
// It starts the RPC server and gossip protocols to handle node
// communication between the cluster.
func New(config Config) *Toystore {
	t := &Toystore{
		ReplicationLevel: config.ReplicationLevel,
		W:                config.W,
		R:                config.R,
		Host:             config.Host,
		RPCPort:          config.RPCPort,
		Ring:             NewRingHead(),
		Data:             config.Store,

		client: NewRpcClient(),
	}

	// Set all logs to show current host
	log.SetFlags(0)
	log.SetPrefix(fmt.Sprintf("[Toystore] %s: ", t.Host))

	// Start new gossip protocol
	t.Members = NewMemberlist(t, config.SeedAddress)

	// Start hinted handoff scan
	t.Hints = NewHintedHandoff(config, t.client)

	// Setup new hash ring
	t.Ring.ReplicationDepth = t.ReplicationLevel
	t.Ring.AddString(t.rpcAddress())

	// Start RPC server
	NewRpcHandler(t)

	return t
}
