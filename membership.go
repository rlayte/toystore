package toystore

import (
	"log"

	"github.com/hashicorp/memberlist"
)

type Member interface {
	Name() string
	Meta() []byte
}

type Members interface {
	Setup(t *Toystore)
	Join(seed string)
	Members() []Member
	Len() int
}

type Memberlist struct {
	list *memberlist.Memberlist
}

type MemberlistNode struct {
	node *memberlist.Node
}

func (m *MemberlistNode) Name() string {
	return m.node.Name
}

func (m *MemberlistNode) Meta() []byte {
	return m.node.Meta
}

func (m *Memberlist) Setup(t *Toystore) {
	memberConfig := memberlist.DefaultLocalConfig()
	memberConfig.BindAddr = t.Host
	memberConfig.Name = t.Host
	memberConfig.IndirectChecks = 0
	memberConfig.Events = &MemberlistEvents{t}

	list, err := memberlist.Create(memberConfig)
	m.list = list

	if err != nil {
		panic("Failed to create memberlist: " + err.Error())
	}
}

func (m *Memberlist) Join(seed string) {
	if seed == "" {
		return
	}

	_, err := m.list.Join([]string{seed})

	if err != nil {
		panic("Failed to join cluster: " + err.Error())
	}
}

func (m *Memberlist) Members() []Member {
	members := []Member{}

	for _, member := range m.list.Members() {
		members = append(members, &MemberlistNode{member})
	}

	return members
}

func (m *Memberlist) Len() int {
	return m.list.NumMembers()
}

type MemberlistEvents struct {
	toystore *Toystore
}

func (m *MemberlistEvents) NotifyJoin(node *memberlist.Node) {
	log.Printf("Toystore joined: %s\n", node)
}

func (m *MemberlistEvents) NotifyLeave(node *memberlist.Node) {
	log.Printf("Toystore left: %s\n", node)
}

func (m *MemberlistEvents) NotifyUpdate(node *memberlist.Node) {
	log.Printf("Toystore update: %s\n", node)
}

func NewMemberlist(t *Toystore, seed string) *Memberlist {
	list := &Memberlist{}
	list.Setup(t)
	list.Join(seed)
	return list
}

func (t *Toystore) Adjacent(address string) bool {
	return t.Ring.Adjacent([]byte(t.rpcAddress()), []byte(address))
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
		key := <-t.requestAddress
		t.receiveAddress <- t.Ring.KeyAddress(key)
	}
}
