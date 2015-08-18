package controllers

import (
	"github.com/gorilla/mux"
	"io"
	"net/http"
)

func GeneratePieHandler(riskEndpoint string) func(http.ResponseWriter, *http.Request) {
	f := func(rw http.ResponseWriter, r *http.Request) {
		idString := mux.Vars(r)["id"]
		piesEndpoint := riskEndpoint + "/pies/" + idString
		resp, _ := http.Get(piesEndpoint)
		io.Copy(rw, resp.Body)
	}
	return f
}
