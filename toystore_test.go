package toystore

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/rlayte/toystore/ring"
	"github.com/rlayte/toystore/store/memory"
)

var numTests = 100
var nodes = []*Toystore{}
var m = &sync.Mutex{}
var hosts = []string{
	"127.0.0.2",
	"127.0.0.3",
	"127.0.0.4",
	"127.0.0.5",
	"127.0.0.6",
}

func node() *Toystore {
	n := nodes[rand.Intn(len(nodes))]
	return n
}

func startNode(host string) {
	log.Println("Starting node", host)
	seedAddress := "127.0.0.2"

	config := Config{
		ReplicationLevel: 3,
		W:                1,
		R:                1,
		RPCPort:          3001,
		Host:             host,
		Store:            memory.New(),
	}

	if host != seedAddress {
		config.SeedAddress = seedAddress
	}

	node := New(config)

	m.Lock()
	nodes = append(nodes, node)
	m.Unlock()
}

func startCluster() {
	log.Println("Starting cluster")

	startNode(hosts[0])

	time.Sleep(time.Second)

	for _, host := range hosts[1:] {
		startNode(host)
	}

	log.Println("Waiting for cluster")
	time.Sleep(time.Second * 15)
	log.Println("Cluster running")
}

func stopCluster() {
}

func randomset(t *testing.T, i int) {
	key := fmt.Sprintf("basic-%d", i)
	value := fmt.Sprintf("basic-value-%d", i)
	n := node()
	n.Put(key, value)
}

func randomget(t *testing.T, i int) {
	key := fmt.Sprintf("basic-%d", i)
	expected := fmt.Sprintf("basic-value-%d", i)
	n := node()
	actual, _ := n.Get(key)

	if actual != expected {
		t.Errorf("%s: %s %s != %s", n.Host, key, actual, expected)
	}
}

func TestIntegration__NoFailures(t *testing.T) {
	var i int

	startCluster()
	defer stopCluster()

	for i = 0; i < numTests; i++ {
		go randomset(t, i)
	}

	time.Sleep(time.Second * 2)

	for i = 0; i < numTests; i++ {
		go randomget(t, i)
	}

	time.Sleep(time.Second)
}

func TestIntegration__NodeJoins(t *testing.T) {
	var i int

	startCluster()
	defer stopCluster()

	for i = 0; i < numTests/2; i++ {
		randomset(t, i)
	}

	// Add new nodes
	startNode("127.0.0.7")
	startNode("127.0.0.8")
	startNode("127.0.0.9")

	log.Println("Wait for new nodes to start")
	time.Sleep(time.Second * 15)

	hosts = append(hosts, "127.0.0.7")
	hosts = append(hosts, "127.0.0.8")
	hosts = append(hosts, "127.0.0.9")

	for i = numTests / 2; i < numTests; i++ {
		go randomset(t, i)
	}

	time.Sleep(time.Second * 2)

	for i = 0; i < numTests; i++ {
		go randomget(t, i)
	}

	time.Sleep(time.Second * 2)
}

type PartitionedRing struct {
	ring ring.Ring
}

func (r *PartitionedRing) Add(key string) {
	// noop
}

func (r *PartitionedRing) ReallyAdd(key string) {
	r.ring.Add(key)
}

func (r *PartitionedRing) Find(key string) string {
	return r.ring.Find(key)
}

func (r *PartitionedRing) FindN(key string, n int) map[string]string {
	return r.ring.FindN(key, n)
}

func (r *PartitionedRing) Fail(key string) {
	r.ring.Fail(key)
}

func (r *PartitionedRing) Revive(key string) {
	r.ring.Revive(key)
}

func (r *PartitionedRing) Adjacent(a, b string) bool {
	return r.ring.Adjacent(a, b)
}

func TestIntegration__Partitions(t *testing.T) {
	startCluster()
	defer stopCluster()

	a := map[string]bool{"127.0.0.3": true, "127.0.0.2": true, "127.0.0.5": true}
	b := map[string]bool{"127.0.0.4": true, "127.0.0.6": true}

	for _, node := range nodes {
		if _, ok := a[node.Host]; ok {
			for host, _ := range b {
				node.Ring.Fail(host)
			}
		}

		if _, ok := b[node.Host]; ok {
			for host, _ := range a {
				node.Ring.Fail(host)
			}
		}
	}

	var i int

	for i = 0; i < numTests; i++ {
		go randomset(t, i)
	}

	time.Sleep(time.Second * 2)

	for i = 0; i < numTests; i++ {
		go randomget(t, i)
	}

	time.Sleep(time.Second)
}
