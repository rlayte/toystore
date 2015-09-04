package toystore

import (
	"time"

	"github.com/hashicorp/memberlist"
)

// Member interface represents an individual node in the cluster.
type Member interface {
	Name() string
	Address() string
	Meta() []byte
}

// MemberlistNode is an implementation of Member that wraps hashicorp's
// memberlist.Node
type MemberlistNode struct {
	node *memberlist.Node
}

// Name returns node.Name
func (m *MemberlistNode) Name() string {
	return m.node.Name
}

// Address returns node.Meta as a string
func (m *MemberlistNode) Address() string {
	return string(m.node.Meta)
}

// Meta returns the raw value of node.Meta
func (m *MemberlistNode) Meta() []byte {
	return m.node.Meta
}

// Members repsents the current nodes in the cluster.
type Members interface {
	Setup(t *Toystore)
	Join(seed string)
	Members() []Member
	Len() int
}

// Memberlist is an implementation of Members using hashicorp's memberlist.
type Memberlist struct {
	list *memberlist.Memberlist
}

// Setup creates a new instance of memberlist, assigns it to list, and
// sets the local nodes meta data as the rpc address.
func (m *Memberlist) Setup(t *Toystore) {
	memberConfig := memberlist.DefaultLocalConfig()
	memberConfig.BindAddr = t.Host
	memberConfig.Name = t.Host
	// Set IndirectChecks to 0 so we see a local view of membership.
	// I.e. we don't care about nodes hidden by partitions.
	memberConfig.IndirectChecks = 0
	// This is set really low for testing purposes. Should be ~100ms.
	memberConfig.GossipInterval = time.Millisecond * 20
	// Sets delegate to handle membership change events.
	memberConfig.Events = &MemberlistEvents{t}

	list, err := memberlist.Create(memberConfig)
	if err != nil {
		panic(err)
	}
	m.list = list
	n := m.list.LocalNode()
	n.Meta = []byte(t.rpcAddress())

	if err != nil {
		panic("Failed to create memberlist: " + err.Error())
	}
}

// Join attempts to join the cluster that the seed node is a member of.
func (m *Memberlist) Join(seed string) {
	if seed == "" {
		return
	}

	_, err := m.list.Join([]string{seed})

	if err != nil {
		panic("Failed to join cluster: " + err.Error())
	}
}

// Members return a list of all current members in the cluster.
func (m *Memberlist) Members() []Member {
	members := []Member{}

	for _, member := range m.list.Members() {
		members = append(members, &MemberlistNode{member})
	}

	return members
}

// Len returns the number of members in the cluster.
func (m *Memberlist) Len() int {
	return m.list.NumMembers()
}

// NewMemberlist returns a new instance of Memberlist, sets up the gossip
// server, and attempts to join the seed node's cluster.
func NewMemberlist(t *Toystore, seed string) *Memberlist {
	list := &Memberlist{}
	list.Setup(t)
	list.Join(seed)
	return list
}

// MemberlistEvents implements memberlist.Events which acts as a delegate for
// membership changes.
type MemberlistEvents struct {
	toystore *Toystore
}

// NotifyJoin is called whenever a new member is discovered and adds it to the
// Toystore instance.
func (m *MemberlistEvents) NotifyJoin(node *memberlist.Node) {
	// Check if the node has an RPC address otherwise we'll try to call
	// empty addresses. This seems to happen as the first event before
	// the local node's Meta data is set.
	if node.Meta != nil {
		member := &MemberlistNode{node}
		m.toystore.AddMember(member)
	}
}

// NotifyLeave is called whenever a node appears to be dead or leaves
// the cluster and then removes the node from the Toystore instance.
func (m *MemberlistEvents) NotifyLeave(node *memberlist.Node) {
	member := &MemberlistNode{node}
	m.toystore.RemoveMember(member)
}

// NotifyUpdate is called whenever an existing node's data is changed.
func (m *MemberlistEvents) NotifyUpdate(node *memberlist.Node) {
	if node.Meta != nil {
		member := &MemberlistNode{node}
		m.toystore.AddMember(member)
	}
}
