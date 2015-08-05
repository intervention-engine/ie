package controllers

import (
	"github.com/gorilla/mux"
	"io"
	"net/http"
)

func PieHandler(rw http.ResponseWriter, r *http.Request) {
	idString := mux.Vars(r)["id"]
	resp, _ := http.Get("http://localhost:9000/pies/" + idString)
	io.Copy(rw, resp.Body)
}
