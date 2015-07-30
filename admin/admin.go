package admin

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/aymerick/raymond"
	"github.com/charlesetc/circle"
	"github.com/julienschmidt/httprouter"
)

var Ring *circle.Circle

func Route(r *httprouter.Router, ring *circle.Circle) {
	r.GET("/", Home())
	r.GET("/favicon.ico", Favicon)
	r.ServeFiles("/static/*filepath", http.Dir("admin/public"))
	fmt.Println(ring)
	Ring = ring
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

		fmt.Println(Ring)
		if Ring == nil {
			ring_string = ""
		} else {
			ring_string = Ring.String()
		}

		context := map[string]string{
			"ring": ring_string,
		}
		output, err := template.Exec(context)
		if err != nil {
			panic(err)
		}
		w.Write([]byte(output))
	}
}
