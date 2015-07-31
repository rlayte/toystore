package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

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

	log.Println("Running server on port", a.store.Port)
	log.Fatal(http.ListenAndServe(a.store.Address(), router))
}

func main() {
	var seed string
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s [port]", os.Args[0])
		os.Exit(1)
	}
	port, err := strconv.Atoi(os.Args[1])

	if err != nil {
		panic(err)
	}

	if port != 3000 {
		seed = ":3010"
	}

	metaData := toystore.ToystoreMetaData{RPCAddress: ":3020"}
	store := toystore.New(port, memory.New(), seed, metaData)
	api := Api{store}

	api.Serve()
}
