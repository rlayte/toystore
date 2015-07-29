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
	Address string
}

type Store interface {
	Get(string) (string, bool)
	Put(string, string)
}

func (t *Toystore) updateMembers() {
	addresses := []string{t.address()}

	for _, member := range t.dive.Members {
		if member.MetaData != nil {
			metaData := member.MetaData.(ToystoreMetaData)
			addresses = append(addresses, metaData.Address)
		}
	}

	t.ring = circle.CircleFromList(addresses)
}

func (t *Toystore) address() string {
	return fmt.Sprintf(":%d", t.port)
}

func (t *Toystore) Get(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	t.updateMembers()

	key := params.ByName("key")
	value, ok := t.data.Get(key)
	lookup := t.ring.KeyAddress([]byte(key))
	address, _ := lookup()

	log.Printf("GET - %s : %s", key, value, address)

	if string(address) != t.address() {
		log.Println("PUT - Forwarding to", string(address))
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

	if string(address) != t.address() {
		log.Println("PUT - Forwarding to", string(address))
	}

	log.Printf("PUT - %s : %s", key, value, address, t.ring)

	t.data.Put(key, value)
}

func (t *Toystore) Serve() {
	router := httprouter.New()

	router.GET("/:key", t.Get)
	router.POST("/:key", t.Put)

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
	n.MetaData = ToystoreMetaData{t.address()}
	gob.RegisterName("ToystoreMetaData", n.MetaData)

	t.dive = n
	return t
}
