package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func Get(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Println("Getting")
}

func Put(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Println("Putting")
}

func main() {
	router := httprouter.New()

	router.GET("/:key", Get)
	router.POST("/:key", Put)

	log.Fatal(http.ListenAndServe(":3000", router))
}
