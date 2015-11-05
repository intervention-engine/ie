package controllers

import (
	"encoding/json"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/utilities"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

type code_request_form struct {
	CodeSystem  string `json:"codesystem"`
	Query       string `json:"query"`
	ResultLimit int    `json:"limit"`
}

func CodeLookup(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	decoder := json.NewDecoder(r.Body)

	var codereqform code_request_form

	err := decoder.Decode(&codereqform)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}

	codesystem, querystring := codereqform.CodeSystem, codereqform.Query
	resultlimit := codereqform.ResultLimit

	codecollection := server.Database.C("codelookup")

	result := []utilities.CodeEntry{}

	iter := codecollection.Find(bson.M{
		"codeSystem": codesystem,
		"$or": []interface{}{
			bson.M{"code": "/" + querystring + "/"},
			bson.M{"name": "/" + querystring + "/"},
		},
	}).Limit(resultlimit).Iter()

	err = iter.All(&result)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(rw).Encode(&result)

}
