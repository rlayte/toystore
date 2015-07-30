package toystore

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/charlesetc/circle"
	"github.com/charlesetc/dive"
	"github.com/julienschmidt/httprouter"
)

type Toystore struct {
	// Config
	ReplicationLevel int
	W                int
	R                int

	// Internal use
	dive *dive.Node
	port int
	data Store
	ring *circle.Circle
}

type ToystoreMetaData struct {
	Address    string
	RPCAddress string
}

type Store interface {
	Get(string) (string, bool)
	Put(string, string)
}

func (t *Toystore) updateMembers() {
	addresses := []string{t.rpcAddress()}

	for _, member := range t.dive.Members {
		if member.MetaData != nil {
			metaData := member.MetaData.(ToystoreMetaData)
			addresses = append(addresses, metaData.RPCAddress)
		}
	}

	t.ring = circle.CircleFromList(addresses)
}

func (t *Toystore) address() string {
	return fmt.Sprintf(":%d", t.port)
}

func (t *Toystore) rpcAddress() string {
	return fmt.Sprintf(":%d", t.port+20)
}

func (t *Toystore) CoordinateGet(key string) (string, bool) {
	t.updateMembers()

	log.Printf("%s coordinating GET request %s.", t.address(), key)

	var value string
	var ok bool

	lookup := t.ring.KeyAddress([]byte(key))

	for address, err := lookup(); err == nil; address, err = lookup() {
		if string(address) != t.rpcAddress() {
			log.Printf("%s sending GET request to %s.", t.address(), address)
		} else {
			log.Printf("Coordinator %s retrieving %s.", t.address(), key)
			value, ok = t.data.Get(key)
		}
	}

	return value, ok
}

func (t *Toystore) Get(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	t.updateMembers()

	key := p.ByName("key")
	lookup := t.ring.KeyAddress([]byte(key))
	address, _ := lookup()

	var value string
	var ok bool

	if string(address) != t.rpcAddress() {
		log.Printf("%s forwarding GET request to %s. %s", t.address(), address, key)
		value, ok = CoordinateGet(string(address), key)
	} else {
		value, ok = t.CoordinateGet(key)
	}

	if !ok {
		w.Header().Set("Status", "404")
		fmt.Fprint(w, "Not found\n")
	} else {
		fmt.Fprint(w, value)
	}
}

func (t *Toystore) CoordinatePut(key string, value string) bool {
	t.updateMembers()

	log.Printf("%s coordinating PUT request %s/%s.", t.address(), key, value)

	lookup := t.ring.KeyAddress([]byte(key))
	writes := 0

	for address, err := lookup(); err == nil; address, err = lookup() {
		if string(address) != t.rpcAddress() {
			log.Printf("%s sending replation request to %s.", t.address(), address)
			ok := PutCall(string(address), key, value)

			if ok {
				writes++
			}
		} else {
			log.Printf("Coordinator %s saving %s/%s.", t.address(), key, value)
			t.data.Put(key, value)
			writes++
		}
	}

	return writes >= t.W
}

func (t *Toystore) Put(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	t.updateMembers()

	key := p.ByName("key")
	value := r.FormValue("data")
	lookup := t.ring.KeyAddress([]byte(key))
	address, _ := lookup()

	var ok bool

	if string(address) != t.rpcAddress() {
		log.Printf("%s forwarding PUT request to coordinator %s.", t.address(), address)
		ok = CoordinatePut(string(address), key, value)
	} else {
		ok = t.CoordinatePut(key, value)
	}

	if ok {
		fmt.Fprint(w, "Success")
	} else {
		fmt.Fprint(w, "Failed")
	}
}

func (t *Toystore) Serve() {
	router := httprouter.New()

	router.GET("/:key", t.Get)
	router.POST("/:key", t.Put)

	go ServeRPC(t)

	log.Println("Running server on port", t.port)
	log.Fatal(http.ListenAndServe(t.address(), router))
}

func New(port int, store Store, seed string, seedMeta interface{}) *Toystore {
	t := &Toystore{
		ReplicationLevel: 3,
		W:                1,
		R:                1,
		port:             port,
		data:             store,
	}

	circle.ReplicationDepth = t.ReplicationLevel

	dive.PingInterval = time.Second
	n := dive.NewNode(port+10, &dive.BasicRecord{Address: seed, MetaData: seedMeta})
	n.MetaData = ToystoreMetaData{t.address(), t.rpcAddress()}
	gob.RegisterName("ToystoreMetaData", n.MetaData)

	t.dive = n
	return t
}
