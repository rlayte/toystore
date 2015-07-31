package toystore

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/aymerick/raymond"
	"github.com/julienschmidt/httprouter"
)

var Toy *Toystore

func AdminRoute(r *httprouter.Router, toystore *Toystore) {
	r.GET("/", Home())
	r.GET("/toystore/force.csv", GraphData)
	r.GET("/favicon.ico", Favicon)
	r.ServeFiles("/static/*filepath", http.Dir("admin/public"))
	// fmt.Println(toystore.ring)
	Toy = toystore
}

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

		fmt.Println(Toy.ring)
		if Toy.ring == nil {
			ring_string = ""
		} else {
			ring_string = Toy.ring.String()
		}

		context := map[string]interface{}{
			"ring": ring_string,
			"keys": Toy.data.Keys(),
		}

		output, err := template.Exec(context)
		if err != nil {
			panic(err)
		}

		w.Write([]byte(output))
	}
}

func GraphData(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if Toy.ring == nil {
		return
	}

	var buf bytes.Buffer
	address_list := Toy.ring.AddressList()
	buf.WriteString("source,target,value\n")
	for i, val := range address_list {
		second_val := address_list[(i+1)%len(address_list)]
		if val != "" && second_val != "" {
			buf.WriteString("localhost") // Tempory hack for d3 parsing.
			buf.WriteString(rpcToAddress(val))
			buf.WriteString(",")
			buf.WriteString("localhost") // Tempory hack for d3 parsing.
			buf.WriteString(rpcToAddress(second_val))
			buf.WriteString(",10\n") // Not sure what value does.
		}
	}

	// Also a little hacky -- connects the ring
	buf.WriteString("localhost") // Tempory hack for d3 parsing.
	buf.WriteString(rpcToAddress(address_list[len(address_list)-1]))
	buf.WriteString(",")
	buf.WriteString("localhost") // Tempory hack for d3 parsing.
	buf.WriteString(rpcToAddress(address_list[1]))
	buf.WriteString(",1\n") // Not sure what value does.

	w.Write(buf.Bytes())
}
