package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/rlayte/dive"
)

type Toystore struct {
	dive *dive.Node
	port int
	data map[string]string
}

func (t *Toystore) Get(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	key := params.ByName("key")
	value, ok := t.data[key]

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
	t.data[key] = value
}

func main() {
	var seed string
	port, err := strconv.Atoi(os.Args[1])

	if port != 3000 {
		seed = ":3010"
	}

	if err != nil {
		panic(err)
	}

	t := Toystore{
		dive: dive.NewNode(port+10, seed),
		port: port,
		data: map[string]string{},
	}

	router := httprouter.New()

	router.GET("/:key", t.Get)
	router.POST("/:key", t.Put)

	log.Println("Running server on port", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), router))
}
