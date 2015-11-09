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

func CodeLookup(rw http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var f code_request_form

	err := decoder.Decode(&f)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}

	codecollection := server.Database.C("codelookup")

	result := []utilities.CodeEntry{}

	query := codecollection.Find(bson.M{
		"codeSystem": f.CodeSystem,
		"$or": []interface{}{
			bson.M{"code": bson.RegEx{Pattern: ".*" + f.Query + ".*", Options: "i"}},
			bson.M{"name": bson.RegEx{Pattern: ".*" + f.Query + ".*", Options: "i"}},
		}}).Limit(f.ResultLimit)

	err = query.All(&result)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(rw).Encode(&result)

}
