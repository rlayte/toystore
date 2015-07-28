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
	port, err := strconv.Atoi(os.Args[1])

	if err != nil {
		panic(err)
	}

	router := httprouter.New()
	dive.NewNode(port+10, "")

	router.GET("/:key", Get)
	router.POST("/:key", Put)

	log.Println("Running server on port", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), router))
}
