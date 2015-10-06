package controllers

import (
	"encoding/json"

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

	intersectRequired := false

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
				intersectRequired = true
			}

			//Encounter
			if coding.System == "http://loinc.org" && coding.Code == "46240-8" {
				equery += "type=" + characteristic.ValueCodeableConcept.Coding[0].System + "|" + characteristic.ValueCodeableConcept.Coding[0].Code + "&"
				intersectRequired = true
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

	if intersectRequired {
		type patientContainer struct {
			ID string `bson:"referenceid"`
		}

		var cResultIDs []struct {
			Patient patientContainer `bson:"patient"`
		}
		cSearchQuery := search.Query{Resource: "Condition", Query: cquery}
		cQ := searcher.CreateQuery(cSearchQuery)
		cQ.Select(bson.M{"patient.referenceid": 1}).All(&cResultIDs)
		cids := make([]string, len(cResultIDs))
		for i := range cResultIDs {
			cids[i] = cResultIDs[i].Patient.ID
		}

		var eResultIDs []struct {
			Patient patientContainer `bson:"patient"`
		}
		eSearchQuery := search.Query{Resource: "Encounter", Query: equery}
		eQ := searcher.CreateQuery(eSearchQuery)
		eQ.Select(bson.M{"patient.referenceid": 1}).All(&eResultIDs)
		eids := make([]string, len(eResultIDs))
		for i := range eResultIDs {
			eids[i] = eResultIDs[i].Patient.ID
		}

		firstIntersect := make(map[string]bool)
		secondIntersect := make(map[string]bool)
		finalIntersect := make(map[string]bool)

		for _, pid := range pids {
			firstIntersect[pid] = true
		}

		for _, cid := range cids {
			if firstIntersect[cid] == true {
				secondIntersect[cid] = true
			}
		}

		for _, eid := range eids {
			if secondIntersect[eid] == true {
				finalIntersect[eid] = true
			}
		}

		intersectPIDs := make([]string, 0, len(finalIntersect))
		for id := range finalIntersect {
			intersectPIDs = append(intersectPIDs, id)
		}
		pids = intersectPIDs
	}

	pCount := len(pids)

	cCollection := server.Database.C("conditions")
	cCount, err := cCollection.Find(bson.M{"patient.referenceid": bson.M{"$in": pids}}).Count()

	eCollection := server.Database.C("encounters")
	eCount, err := eCollection.Find(bson.M{"patient.referenceid": bson.M{"$in": pids}}).Count()

	newResultMap := map[string]int{
		"patients":   pCount,
		"conditions": cCount,
		"encounters": eCount,
	}

	json.NewEncoder(rw).Encode(newResultMap)
}
