// THIS FILE IS SAFE TO EDIT. It will not be overwritten when rerunning go-raml.
package api

import (
	"log"
	"net/http"

	"github.com/vennekilde/gw2verify/internal/api/goraml"

	"github.com/gorilla/mux"
	"gopkg.in/validator.v2"
)

func StartServer() {
	// input validator
	validator.SetValidationFunc("multipleOf", goraml.MultipleOf)

	r := mux.NewRouter()

	// home page
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	initRoutes(r)

	log.Println("starting server")
	http.ListenAndServe(":5000", r)
}
