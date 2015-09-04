package toystore

import (
	"time"

	"github.com/rlayte/toystore/store"
)

// Config defines the variables used for a Toystore node.
type Config struct {
	// Number of nodes to store a key on.
	ReplicationLevel int

	// W is the number of times a key must be replicated for a Put operation
	// to be successful.
	W int

	// R is the number of nodes a key must be read from for a Get operation
	// to be successful.
	R int

	// RPCPort is the port Toystore will use for RPC between nodes.
	RPCPort int

	// GossipPort is the port Toystore will use for membership updates via
	// its gossip protocol.
	GossipPort int

	// Host is the ip Toystore will bind to.
	Host string

	// SeedAddress is a known Toystore node that used to initial join the
	// cluster.
	SeedAddress string

	// Store is an implementation of the Store interface the handles persisting
	// data.
	Store store.Store

	// HandoffInterval is the time between scans of the hinted handoff list.
	HandoffInterval time.Duration
}
