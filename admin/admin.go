package admin

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/aymerick/raymond"
	"github.com/julienschmidt/httprouter"
	"github.com/rlayte/toystore"
)

var Toy *toystore.Toystore

func Favicon(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	output, err := ioutil.ReadFile("admin/public/favicon.ico")
	if err != nil {
		panic(err)
	}
	w.Write(output)
}

func Home() func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	contents, err := ioutil.ReadFile("admin/views/home.html")
	if err != nil {
		panic(err)
	}
	template, err := raymond.Parse(string(contents))
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

		var ring_string string

		fmt.Println(Toy.Ring)
		if Toy.Ring == nil {
			ring_string = ""
		} else {
			ring_string = Toy.Ring.String()
		}

		context := map[string]interface{}{
			"ring": ring_string,
			"keys": Toy.Data.Keys(),
		}

		output, err := template.Exec(context)
		if err != nil {
			panic(err)
		}

		w.Write([]byte(output))
	}
}

func Get(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	Toy.UpdateMembers()
	key := r.FormValue("key")

	value, ok := Toy.Get(key)

	if !ok {
		w.Header().Set("Status", "404")
		fmt.Fprint(w, "Not found\n")
		return
	} else {
		fmt.Fprint(w, value)
	}
}

func Put(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	Toy.UpdateMembers()
	key := r.FormValue("key")
	value := r.FormValue("data")

	Toy.Put(key, value)

	http.Redirect(w, r, "/", 301)
}

func Serve(t *toystore.Toystore) {
	Toy = t
	router := httprouter.New()

	router.GET("/", Home())
	router.GET("/toystore/force.csv", GraphData)
	router.GET("/favicon.ico", Favicon)
	router.ServeFiles("/static/*filepath", http.Dir("admin/public"))
	router.GET("/api", Get)
	router.POST("/api", Put)

	go toystore.ServeRPC(t)

	log.Println("Running server on port", t.Port)
	log.Fatal(http.ListenAndServe(t.Address(), router))
}

func GraphData(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if Toy.Ring == nil {
		return
	}

	var buf bytes.Buffer
	address_list := Toy.Ring.AddressList()
	buf.WriteString("source,target,value\n")
	for i, val := range address_list {
		second_val := address_list[(i+1)%len(address_list)]
		if val != "" && second_val != "" {
			buf.WriteString("localhost") // Tempory hack for d3 parsing.
			buf.WriteString(toystore.RpcToAddress(val))
			buf.WriteString(",")
			buf.WriteString("localhost") // Tempory hack for d3 parsing.
			buf.WriteString(toystore.RpcToAddress(second_val))
			buf.WriteString(",10\n") // Not sure what value does.
		}
	}

	// Also a little hacky -- connects the ring
	buf.WriteString("localhost") // Tempory hack for d3 parsing.
	buf.WriteString(toystore.RpcToAddress(address_list[len(address_list)-1]))
	buf.WriteString(",")
	buf.WriteString("localhost") // Tempory hack for d3 parsing.
	buf.WriteString(toystore.RpcToAddress(address_list[1]))
	buf.WriteString(",1\n") // Not sure what value does.

	w.Write(buf.Bytes())
}
