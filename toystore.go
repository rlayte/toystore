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

func (t *Toystore) Get(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	t.updateMembers()

	key := params.ByName("key")
	lookup := t.ring.KeyAddress([]byte(key))
	address, _ := lookup()
	var value string
	var ok bool

	if string(address) != t.rpcAddress() {
		log.Printf("%s forwarding GET request to %s. %s", t.address(), address, key)
		value, ok = GetCall(string(address), key)
	} else {
		log.Printf("%s handling GET request. %s", t.address(), key)
		value, ok = t.data.Get(key)
	}

	if !ok {
		w.Header().Set("Status", "404")
		fmt.Fprint(w, "Not found")
		return
	} else {
		fmt.Fprint(w, value)
	}
}

func (t *Toystore) Put(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	t.updateMembers()

	key := params.ByName("key")
	value := r.FormValue("data")
	lookup := t.ring.KeyAddress([]byte(key))
	address, _ := lookup()

	if string(address) != t.rpcAddress() {
		log.Printf("%s forwarding PUT request to %s. %s:%s", t.address(), address, key, value)
		PutCall(string(address), key, value)
	} else {
		log.Printf("%s handling GET request. %s:%s", t.address(), key, value)
		t.data.Put(key, value)
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
		port: port,
		data: store,
	}

	dive.PingInterval = time.Second
	n := dive.NewNode(port+10, &dive.BasicRecord{Address: seed, MetaData: seedMeta})
	n.MetaData = ToystoreMetaData{t.address(), t.rpcAddress()}
	gob.RegisterName("ToystoreMetaData", n.MetaData)

	t.dive = n
	return t
}
