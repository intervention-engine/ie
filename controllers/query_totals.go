package controllers

import (
	"encoding/json"

	"github.com/gorilla/mux"
	fhirmodels "github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/models"
	"gopkg.in/mgo.v2"
	"net/http"
)

func pipelineExecutor(rw http.ResponseWriter, r *http.Request, pp models.PipelineProducer, pipelineType string) {
	query, err := server.LoadQuery(r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	pipeline := pp(query)
	switch pipelineType {
	case "patient":
		qr, err := pipeline.ExecutePatientList(server.Database)
		if err != nil && err != mgo.ErrNotFound {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(rw).Encode(qr)
	case "count":
		qr, err := pipeline.ExecuteCount(server.Database)
		if err != nil && err != mgo.ErrNotFound {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(rw).Encode(qr)
	}
}

func ConditionTotalHandler(rw http.ResponseWriter, r *http.Request) {
	pipelineExecutor(rw, r, models.NewConditionPipeline, "count")
}

func EncounterTotalHandler(rw http.ResponseWriter, r *http.Request) {
	pipelineExecutor(rw, r, models.NewEncounterPipeline, "count")
}

func PatientListHandler(rw http.ResponseWriter, r *http.Request) {
	pipelineExecutor(rw, r, models.NewPipeline, "patient")
}

func InstaCountHandler(rw http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	query := &fhirmodels.Query{}
	err := decoder.Decode(query)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
	queryType := mux.Vars(r)["type"]
	var pipeline models.Pipeline
	switch queryType {
	case "patient":
		pipeline = models.NewPipeline(query)
	case "encounter":
		pipeline = models.NewEncounterPipeline(query)
	case "condition":
		pipeline = models.NewConditionPipeline(query)
	}
	qr, err := pipeline.ExecuteCount(server.Database)
	if err != nil && err != mgo.ErrNotFound {
		http.Error(rw, err.Error(), http.StatusTeapot)
		return
	}
	json.NewEncoder(rw).Encode(qr)
}
