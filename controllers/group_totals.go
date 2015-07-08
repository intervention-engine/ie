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
	group, err := server.LoadGroup(r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	pipeline := pp(group)
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
	group := &fhirmodels.Group{}
	err := decoder.Decode(group)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	queryType := mux.Vars(r)["type"]
	var pipeline models.Pipeline
	switch queryType {
	case "patient":
		pipeline = models.NewPipeline(group)
	case "encounter":
		pipeline = models.NewEncounterPipeline(group)
	case "condition":
		pipeline = models.NewConditionPipeline(group)
	}
	qr, err := pipeline.ExecuteCount(server.Database)
	if err != nil && err != mgo.ErrNotFound {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(rw).Encode(qr)
}

func InstaCountAllHandler(rw http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	group := &fhirmodels.Group{}
	err := decoder.Decode(group)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	pipelineMap := make(map[string]models.Pipeline)

	pipelineMap["patients"] = models.NewPipeline(group)
	pipelineMap["encounters"] = models.NewEncounterPipeline(group)
	pipelineMap["conditions"] = models.NewConditionPipeline(group)

	resultMap := make(map[string]int)

	for pipelineType, pipeline := range pipelineMap {
		qr, err := pipeline.ExecuteCount(server.Database)
		if err != nil && err != mgo.ErrNotFound {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		resultMap[pipelineType] = qr.Total
	}

	json.NewEncoder(rw).Encode(resultMap)
}
