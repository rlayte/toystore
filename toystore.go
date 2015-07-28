package toystore

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/rlayte/dive"
)

type Toystore struct {
	dive *dive.Node
	port int
	data Store
}

type Store interface {
	Get(string) (string, bool)
	Put(string, string)
}

func (t *Toystore) Get(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	key := params.ByName("key")
	value, ok := t.data.Get(key)

	log.Printf("GET - %s : %s", key, value, t.dive.Members)

	if !ok {
		w.Header().Set("Status", "404")
		fmt.Fprint(w, "Not found")
		return
	} else {
		fmt.Fprint(w, value)
	}
}

func (t *Toystore) Put(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	key := params.ByName("key")
	value := r.FormValue("data")
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
	return &Toystore{
		dive: dive.NewNode(port+10, seed),
		port: port,
		data: store,
	}
}
