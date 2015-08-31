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

	log.Println("Running server on", a.store.Address())
	log.Fatal(http.ListenAndServe(a.store.Address(), router))
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
		ClientPort:       3000,
		RPCPort:          3001,
		Host:             host,
		Store:            memory.New(),
	}

	if host != seed {
		config.SeedAddress = seed
	}

	seedRPCAddress := fmt.Sprintf("%s:%d", seed, config.RPCPort)
	metaData := toystore.ToystoreMetaData{RPCAddress: seedRPCAddress}
	store := toystore.New(config, metaData)
	api := Api{store}

	api.Serve()
}
