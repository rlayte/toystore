// Package toystore is a simple implementation of a Dynamo database -
// http://www.allthingsdistributed.com/files/amazon-dynamo-sosp2007.pdf
//
// For more information see the project's readme -
// https://github.com/rlayte/toystore
package toystore

import (
	"fmt"
	"log"
	"os"

	"github.com/rlayte/toystore/data"
	"github.com/rlayte/toystore/ring"
	"github.com/rlayte/toystore/store"
)

// Toystore represents an individual node in a Toystore cluster.
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
	Data store.Store

	// Hash ring for nodes in the cluster.
	Ring ring.Ring

	// Concrete Members implementation to represent the current cluster state.
	Members Members

	// Store of hinted data meant for other nodes.
	Hints *HintedHandoff

	// Concrete PeerClient implementation to make calls to other nodes.
	client PeerClient

	// Concrete Transferrer implementation to transfer data to other nodes.
	transferrer Transferrer

	// Custom logger. Format: [Toystore] {host}: {statement}
	log *log.Logger
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
	address := t.Ring.Find(key)
	var data *data.Data
	var ok bool

	if t.isCoordinator(address) {
		data, ok = t.CoordinateGet(key)
	} else {
		t.log.Printf("Forwarding GET request to %s for %s", address, key)
		data, ok = t.client.CoordinateGet(address, key)
	}

	if ok {
		return data.Value, ok
	}

	return nil, ok
}

// Put finds the key on the correct node in the cluster, sets
// the value and returns a status bool.
// If the key is owned by current node then it coordinates the operation.
// Otherwise it sends the coordination request to the correct node.
func (t *Toystore) Put(key string, value interface{}) (ok bool) {
	address := t.Ring.Find(key)

	if t.isCoordinator(address) {
		ok = t.CoordinatePut(data.New(key, value))
	} else {
		t.log.Printf("Forwarding PUT request to coordinator %s for %s", address, value)
		ok = t.client.CoordinatePut(address, data.New(key, value))
	}
	return
}

// GetString returns a string of the value for the specified key/value pair.
func (t *Toystore) GetString(key string) (string, bool) {
	d, ok := t.Get(key)
	return d.(string), ok
}

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
	t.log.Printf("Coordinating GET request %s.", key)

	values := []*data.Data{}
	nodes := t.Ring.FindN(key, t.ReplicationLevel)
	reads := 0

	for _, address := range nodes {
		if address != t.rpcAddress() {
			t.log.Printf("GET request to %s for %s", address, key)
			value, ok := t.client.Get(address, key)

			if ok {
				values = append(values, value)
				reads++
			}
		} else {
			t.log.Printf("Coordinator retrieving %s", key)
			value, ok := t.Data.Get(key)

			if ok {
				values = append(values, value)
				reads++
			}
		}
	}

	// Add the newest value found to the local database
	for _, value := range values {
		t.Merge(value)
	}

	value, _ := t.Data.Get(key)

	return value, reads >= t.R
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
	t.log.Printf("Coordinating PUT request %v", value)

	nodes := t.Ring.FindN(key, t.ReplicationLevel)
	writes := 0

	for address, hint := range nodes {
		if address != t.rpcAddress() {
			var ok bool

			if hint != address {
				t.log.Printf("Sending hint to %s for %s (%s)", address, hint, value)
				ok = t.client.HintPut(address, string(hint), value)
			} else {
				t.log.Printf("PUT request to %s for %v", address, value)
				ok = t.client.Put(address, value)
			}

			if ok {
				writes++
			}
		} else {
			t.log.Printf("Coordinator saving %s", value)
			ok := t.Data.Put(value)

			if ok {
				writes++
			}
		}
	}

	return writes >= t.W
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
	keys := t.Data.Keys()
	items := []*data.Data{}

	for _, key := range keys {
		val, _ := t.Data.Get(key)

		if t.Ring.Find(key) == address {
			items = append(items, val)
		}
	}

	if len(items) > 0 {
		t.log.Printf("Transferring to %s. Ring: %s", address, t.Ring, items)
		t.transferrer.Transfer(address, items)
	}
}

// AddMember adds a new node to the hash ring.
// If the new node is adjacent to the current node then it transfers
// any keys in its range that should be owned by the new node.
func (t *Toystore) AddMember(member Member) {
	t.log.Printf("Adding member %s", member.Name())
	t.Ring.Add(member.Address())
	localAddress := t.rpcAddress()
	adjacent := t.Ring.Adjacent(member.Address(), localAddress)

	if adjacent {
		t.Transfer(member.Address())
	}
}

// RemoveMember removes a member from the hash ring.
func (t *Toystore) RemoveMember(member Member) {
	if member.Address() != t.rpcAddress() {
		t.log.Printf("Removing member %s", member.Name())
		t.Ring.Fail(member.Address())
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
		Ring:             ring.NewHashRing(),
		Data:             config.Store,
	}

	// Set all logs to show current host
	prefix := fmt.Sprintf("[Toystore] %s: ", t.Host)
	t.log = log.New(os.Stderr, prefix, 0)

	// Initialize RPC client for inter-node communication
	client := NewRpcClient()
	t.client = client
	t.transferrer = client

	// Start new gossip protocol
	t.Members = NewMemberlist(t, config.SeedAddress)

	// Start hinted handoff scan
	t.Hints = NewHintedHandoff(config, client)

	// Setup new hash ring
	t.Ring.Add(t.rpcAddress())

	// Start RPC server
	NewRpcHandler(t)

	return t
}
