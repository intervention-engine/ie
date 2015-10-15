package controllers

import (
	"encoding/json"

	"net/http"
	"strings"
	"time"

	fhirmodels "github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/search"
	"github.com/intervention-engine/fhir/server"
	"gopkg.in/mgo.v2/bson"
)

func groupListResolver(group fhirmodels.Group) []string {
	pquery := ""
	cquery := ""
	equery := ""

	hasConditionCriteria := false
	hasEncounterCriteria := false

	for _, characteristic := range group.Characteristic {

		codings := characteristic.Code.Coding

		for _, coding := range codings {
			//Age
			if coding.System == "http://loinc.org" && coding.Code == "21612-7" {
				highAgeDate := time.Date(time.Now().Year()-int(*characteristic.ValueRange.Low.Value), time.Now().Month(), time.Now().Day(), 23, 59, 59, 999*int(time.Millisecond), time.UTC).Format("2006-01-02T15:04:05.999")
				lowAgeDate := time.Date(time.Now().Year()-int(*characteristic.ValueRange.High.Value), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC).Format("2006-01-02T15:04:05.999")
				pquery += "birthdate=lte" + highAgeDate + "&birthdate=gte" + lowAgeDate + "&"
				cquery += "patient.birthdate=lte" + highAgeDate + "&patient.birthdate=gte" + lowAgeDate + "&"
				equery += "patient.birthdate=lte" + highAgeDate + "&patient.birthdate=gte" + lowAgeDate + "&"
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
				hasConditionCriteria = true
			}

			//Encounter
			if coding.System == "http://loinc.org" && coding.Code == "46240-8" {
				equery += "type=" + characteristic.ValueCodeableConcept.Coding[0].System + "|" + characteristic.ValueCodeableConcept.Coding[0].Code + "&"
				hasEncounterCriteria = true
			}
		}
	}

	pquery = trimSuffix(pquery, "&")
	cquery = trimSuffix(cquery, "&")
	equery = trimSuffix(equery, "&")

	searcher := search.NewMongoSearcher(server.Database)

	var pids []string
	var cids []string
	var eids []string

	type patientContainer struct {
		ID string `bson:"referenceid"`
	}

	//We only need to query for conditions if the group contains a condition criteria
	if hasConditionCriteria {
		var cResultIDs []struct {
			Patient patientContainer `bson:"patient"`
		}
		cSearchQuery := search.Query{Resource: "Condition", Query: cquery}
		cQ := searcher.CreateQuery(cSearchQuery)
		cQ.Select(bson.M{"patient.referenceid": 1}).All(&cResultIDs)
		cids = make([]string, len(cResultIDs))
		for i := range cResultIDs {
			cids[i] = cResultIDs[i].Patient.ID
		}
	}

	//We only need to query for encounters if the group contains an encounter criteria
	if hasEncounterCriteria {
		var eResultIDs []struct {
			Patient patientContainer `bson:"patient"`
		}
		eSearchQuery := search.Query{Resource: "Encounter", Query: equery}
		eQ := searcher.CreateQuery(eSearchQuery)
		eQ.Select(bson.M{"patient.referenceid": 1}).All(&eResultIDs)
		eids = make([]string, len(eResultIDs))
		for i := range eResultIDs {
			eids[i] = eResultIDs[i].Patient.ID
		}
	}

	//If we have both a condition critera and an encounter criteria, the patient ID list is the intersection of the resultant IDs from those queries
	if hasConditionCriteria && hasEncounterCriteria {
		cidMap := make(map[string]bool)
		intersectMap := make(map[string]bool)

		for _, cid := range cids {
			cidMap[cid] = true
		}

		for _, eid := range eids {
			if cidMap[eid] == true {
				intersectMap[eid] = true
			}
		}

		intersectPIDs := make([]string, 0, len(intersectMap))
		for id := range intersectMap {
			intersectPIDs = append(intersectPIDs, id)
		}
		pids = intersectPIDs
	} else if hasConditionCriteria {
		//If we only have a condition criteria, the patient ID list is the resultant IDs from that query
		pids = cids
	} else if hasEncounterCriteria {
		//If we only have an encounter criteria, the patient ID list is the resultant IDs from that query
		pids = eids
	} else {
		//If we have neither a condition nor encounter criteria, the patient ID list is the result of the demographic query against the patient collection
		var pResultIDs []struct {
			ID string `bson:"_id"`
		}
		pSearchQuery := search.Query{Resource: "Patient", Query: pquery}
		pQ := searcher.CreateQuery(pSearchQuery)
		pQ.Select(bson.M{"_id": 1}).All(&pResultIDs)
		pids = make([]string, len(pResultIDs))
		for i := range pResultIDs {
			pids[i] = pResultIDs[i].ID
		}
	}

	return pids
}

func PatientListHandler(rw http.ResponseWriter, r *http.Request) {
	groupController := server.ResourceController{Name: "Group"}
	group, err := groupController.LoadResource(r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
	patientIds := groupListResolver(*group.(*fhirmodels.Group))
	responseMap := map[string][]string{
		"patientids": patientIds,
	}
	json.NewEncoder(rw).Encode(responseMap)
}

func trimSuffix(s, suffix string) string {
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

	pids := groupListResolver(*group)
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
