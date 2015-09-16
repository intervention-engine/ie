package controllers

import (
	"encoding/json"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	fhirmodels "github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/search"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strings"
	"time"
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

func TrimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}

func InstaCountAllHandler(rw http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	group := &fhirmodels.Group{}
	err := decoder.Decode(group)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	pquery := ""
	cquery := ""
	equery := ""

	for _, characteristic := range group.Characteristic {

		codings := characteristic.Code.Coding

		for _, coding := range codings {
			//Age
			if coding.System == "http://loinc.org" && coding.Code == "21612-7" {
				highAgeDate := time.Date(time.Now().Year()-int(*characteristic.ValueRange.Low.Value), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC).Format("2006-01-02")
				lowAgeDate := time.Date(time.Now().Year()-int(*characteristic.ValueRange.High.Value), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC).Format("2006-01-02")
				pquery += "birthdate=lt" + highAgeDate + "&birthdate=gt" + lowAgeDate + "&"
				cquery += "patient.birthdate=lt" + highAgeDate + "&patient.birthdate=gt" + lowAgeDate + "&"
				equery += "patient.birthdate=lt" + highAgeDate + "&patient.birthdate=gt" + lowAgeDate + "&"
			}

			//Gender
			if coding.System == "http://loinc.org" && coding.Code == "21840-4" {
				gender := characteristic.ValueCodeableConcept.Coding[0].Code
				pquery += "gender=" + gender + "&"
				cquery += "patient.gender=" + gender + "&"
				equery += "patient.gender=" + gender + "&"
			}

			//Condition
			if coding.System == "http://loinc.org" && coding.Code == "11450-4" {
				cquery += "code=" + characteristic.ValueCodeableConcept.Coding[0].System + "|" + characteristic.ValueCodeableConcept.Coding[0].Code + "&"
			}

			//Encounter
			if coding.System == "http://loinc.org" && coding.Code == "46240-8" {
				equery += "type=" + characteristic.ValueCodeableConcept.Coding[0].System + "|" + characteristic.ValueCodeableConcept.Coding[0].Code + "&"
			}
		}
	}

	pquery = TrimSuffix(pquery, "&")
	cquery = TrimSuffix(cquery, "&")
	equery = TrimSuffix(equery, "&")

	searcher := search.NewMongoSearcher(server.Database)

	var pResultIDs []struct {
		ID string `bson:"_id"`
	}
	pSearchQuery := search.Query{Resource: "Patient", Query: pquery}
	pQ := searcher.CreateQuery(pSearchQuery)
	pQ.Select(bson.M{"_id": 1}).All(&pResultIDs)
	pids := make([]string, len(pResultIDs))
	for i := range pResultIDs {
		pids[i] = pResultIDs[i].ID
	}

	var cResultIDs []struct {
		ID string `bson:"patient.referenceid"`
	}
	cSearchQuery := search.Query{Resource: "Condition", Query: cquery}
	cQ := searcher.CreateQuery(cSearchQuery)
	cQ.Select(bson.M{"patient.referenceid": 1}).All(&cResultIDs)

	cids := make([]string, len(cResultIDs))
	for i := range cResultIDs {
		cids[i] = cResultIDs[i].ID
	}

	var eResultIDs []struct {
		ID string `bson:"patient.referenceid"`
	}
	eSearchQuery := search.Query{Resource: "Encounter", Query: equery}
	eQ := searcher.CreateQuery(eSearchQuery)
	eQ.Select(bson.M{"patient.referenceid": 1}).All(&eResultIDs)
	eids := make([]string, len(eResultIDs))
	for i := range eResultIDs {
		eids[i] = eResultIDs[i].ID
	}
	spew.Dump(len(pids), len(cids), len(eids))

	intersectMap := make(map[string]bool)

	for _, pid := range pids {
		if pid != "" {
			intersectMap[pid] = true
		}
	}

	for _, cid := range cids {
		if cid != "" {
			intersectMap[cid] = true
		}
	}

	for _, eid := range eids {
		if eid != "" {
			intersectMap[eid] = true
		}
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
