// api/http is an example application of Toystore.
// It wraps the library in a simple http API that accepts and
// returns json.
//
// Usage:
//  # Start the seed
//  go run api/http.go 127.0.0.2
//
//  # Start other nodes
//  go run api/http.go 127.0.0.3/n
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/rlayte/toystore"
	"github.com/rlayte/toystore/adapters/memory"
)

type Api struct {
	store *toystore.Toystore
}

func (a *Api) Meta(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	fmt.Fprint(w, "Node meta data\n")
}

func (a *Api) Address() string {
	return fmt.Sprintf("%s:%d", a.store.Host, 3000)
}

func (a *Api) Get(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	key := params.ByName("key")
	value, ok := a.store.Get(key)

	if !ok {
		w.Header().Set("Status", "404")
		fmt.Fprint(w, "Not found\n")
		return
	} else {
		fmt.Fprint(w, value)
	}
}

func (a *Api) Put(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	key := r.FormValue("key")
	value := r.FormValue("value")

	ok := a.store.Put(key, value)

	if ok {
		fmt.Fprint(w, "Success\n")
	} else {
		fmt.Fprint(w, "Failed\n")
	}
}

func (a *Api) Serve() {
	router := httprouter.New()

	router.GET("/", a.Meta)
	router.GET("/:key", a.Get)
	router.POST("/", a.Put)

	log.Println("Running server on", a.Address())
	log.Fatal(http.ListenAndServe(a.Address(), router))
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s [host]", os.Args[0])
		os.Exit(1)
	}

	seed := "127.0.0.2"
	host := os.Args[1]

	config := toystore.Config{
		ReplicationLevel: 3,
		W:                1,
		R:                1,
		RPCPort:          3001,
		Host:             host,
		Store:            memory.New(),
	}

	if host != seed {
		config.SeedAddress = seed
	}

	store := toystore.New(config)
	api := Api{store}

	api.Serve()
}
