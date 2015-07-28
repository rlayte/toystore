package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

var data map[string]string = map[string]string{}

func Get(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	key := params.ByName("key")
	value, ok := data[key]

	if !ok {
		w.Header().Set("Status", "404")
		fmt.Fprint(w, "Not found")
		return
	} else {
		fmt.Fprint(w, value)
	}
}

func Put(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	key := params.ByName("key")
	value := r.FormValue("data")
	data[key] = value
}

func main() {
	router := httprouter.New()

	router.GET("/:key", Get)
	router.POST("/:key", Put)

	log.Println("Running server on port 3000")
	log.Fatal(http.ListenAndServe(":3000", router))
}
