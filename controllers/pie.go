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
		resp, err := http.Get(piesEndpoint)

		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}

		io.Copy(rw, resp.Body)
	}
	return f
}
