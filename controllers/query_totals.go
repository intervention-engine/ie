package controllers

import (
	"encoding/json"

	"net/http"

	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/models"
)

func pipelineExecutor(rw http.ResponseWriter, r *http.Request, pp models.PipelineProducer) {
	query, err := server.LoadQuery(r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	pipeline := pp(query)
	qr, err := pipeline.ExecuteCount(server.Database)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(rw).Encode(qr)
}

func ConditionTotalHandler(rw http.ResponseWriter, r *http.Request) {
	pipelineExecutor(rw, r, models.NewConditionPipeline)
}

func EncounterTotalHandler(rw http.ResponseWriter, r *http.Request) {
	pipelineExecutor(rw, r, models.NewEncounterPipeline)
}
