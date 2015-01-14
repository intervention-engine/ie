package controllers

import (
	"encoding/json"

	"net/http"

	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/middleware"
	"github.com/intervention-engine/ie/models"
)

func ConditionTotalHandler(rw http.ResponseWriter, r *http.Request) {
	query, err := server.LoadQuery(r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	pipeline := models.CreateConditionPipeline(query)
	factCollection := server.Database.C("facts")
	qr := &middleware.QueryResult{}
	factCollection.Pipe(pipeline).One(qr)
	json.NewEncoder(rw).Encode(qr)
}

func EncounterTotalHandler(rw http.ResponseWriter, r *http.Request) {
	query, err := server.LoadQuery(r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	pipeline := models.CreateEncounterPipeline(query)
	factCollection := server.Database.C("facts")
	qr := &middleware.QueryResult{}
	factCollection.Pipe(pipeline).One(qr)
	json.NewEncoder(rw).Encode(qr)
}
