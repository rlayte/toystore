package toystore

import (
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

type Store interface {
	Get(string) (string, bool)
	Put(string, string)
}

func (t *Toystore) updateMembers() {
	addresses := []string{t.dive.Address()}

	for _, member := range t.dive.Members {
		addresses = append(addresses, member.Address)
	}

	t.ring = circle.CircleFromList(addresses)
}

func (t *Toystore) Get(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	t.updateMembers()

	key := params.ByName("key")
	value, ok := t.data.Get(key)
	lookup := t.ring.KeyAddress([]byte(key))
	address, _ := lookup()

	log.Printf("GET - %s : %s", key, value)
	log.Println("Key address", string(address), t.ring)

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

	log.Printf("PUT - %s : %s", key, value)

	t.data.Put(key, value)
}

func (t *Toystore) Serve() {
	router := httprouter.New()

	router.GET("/:key", t.Get)
	router.POST("/:key", t.Put)

	log.Println("Running server on port", t.port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", t.port), router))
}

func New(port int, store Store, seed string) *Toystore {
	dive.PingInterval = time.Second

	return &Toystore{
		dive: dive.NewNode(port+10, seed),
		port: port,
		data: store,
	}
}
