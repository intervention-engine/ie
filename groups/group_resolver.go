package groups

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	fhir "github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/search"
	"gopkg.in/mgo.v2/bson"
)

func resolveGroup(cInfo *CharacteristicInfo, searcher *search.MongoSearcher) (pids []string, err error) {
	cMap := make(map[string]bool)
	eMap := make(map[string]bool)

	type patientContainer struct {
		ID string `bson:"referenceid"`
	}

	// We only need to query for conditions if the group contains a condition criteria
	// NOTE: We can't prefilter unconfirmed conditions (FHIR doesn't have a search parameter for
	// verificationStatus) and our API doesn't let us insert non-FHIR criteria, so we must
	// post-process the results to filter them out.
	if cInfo.HasConditionCharacteristics {
		var cResults []struct {
			Patient patientContainer `bson:"patient"`
			Status  string           `bson:"verificationStatus"`
		}
		cSearchQuery := search.Query{Resource: "Condition", Query: cInfo.ConditionQueryValues.Encode()}
		cQ := searcher.CreateQueryWithoutOptions(cSearchQuery)

		if err := cQ.Select(bson.M{"patient.referenceid": 1, "verificationStatus": 1}).All(&cResults); err != nil {
			return nil, err
		}
		for i := range cResults {
			if cResults[i].Status == "confirmed" {
				cMap[cResults[i].Patient.ID] = true
			}
		}
	}

	// We only need to query for encounters if the group contains an encounter criteria
	if cInfo.HasEncounterCharacteristics {
		var eResults []struct {
			Patient patientContainer `bson:"patient"`
		}
		eSearchQuery := search.Query{Resource: "Encounter", Query: cInfo.EncounterQueryValues.Encode()}
		eQ := searcher.CreateQueryWithoutOptions(eSearchQuery)
		if err := eQ.Select(bson.M{"patient.referenceid": 1}).All(&eResults); err != nil {
			return nil, err
		}
		for i := range eResults {
			eMap[eResults[i].Patient.ID] = true
		}
	}

	// Now figure out the patient list, which can come in one of four ways:
	// (1) Group has condition AND encounter characteristics (+ optional demographics) --> patients who have matching conditions AND encounters
	// (2) Group has only condition characteristics (+ optional demographics) --> patients who have matching conditions
	// (3) Group has only encounter characteristics (+ optional demographics) --> patients who have matching encounters
	// (4) Group has only demographics --> patients who match demographics
	if cInfo.HasConditionCharacteristics && cInfo.HasEncounterCharacteristics {
		// Only count patients with matching conditions and encounters
		for pid := range cMap {
			if _, ok := eMap[pid]; ok {
				pids = append(pids, pid)
			}
		}
	} else if cInfo.HasConditionCharacteristics {
		// Only count patients with matching conditions
		pids = make([]string, 0, len(cMap))
		for pid := range cMap {
			pids = append(pids, pid)
		}
	} else if cInfo.HasEncounterCharacteristics {
		// Only count patients with matching encounters
		pids = make([]string, 0, len(eMap))
		for pid := range eMap {
			pids = append(pids, pid)
		}
	} else {
		// Only count patients with matching demographics
		var resultIDs []struct {
			ID string `bson:"_id"`
		}
		q := searcher.CreateQueryWithoutOptions(search.Query{Resource: "Patient", Query: cInfo.PatientQueryValues.Encode()})
		if err := q.Select(bson.M{"_id": 1}).All(&resultIDs); err != nil {
			return nil, err
		}
		pids = make([]string, len(resultIDs))
		for i := range resultIDs {
			pids[i] = resultIDs[i].ID
		}
	}

	return pids, nil
}

func resolveGroupCounts(cInfo *CharacteristicInfo, searcher *search.MongoSearcher) (patients, conditions, encounters int, err error) {
	pids, err := resolveGroup(cInfo, searcher)
	if err != nil {
		return 0, 0, 0, err
	}
	patients = len(pids)

	cQuery := bson.M{
		"verificationStatus": "confirmed",
		"patient.referenceid": bson.M{
			"$in": pids,
		},
	}
	conditions, err = searcher.GetDB().C("conditions").Find(cQuery).Count()
	if err != nil {
		return 0, 0, 0, err
	}

	eQuery := bson.M{
		"patient.referenceid": bson.M{
			"$in": pids,
		},
	}
	encounters, err = searcher.GetDB().C("encounters").Find(eQuery).Count()
	if err != nil {
		return 0, 0, 0, err
	}

	return patients, conditions, encounters, nil
}

type CharacteristicInfo struct {
	PatientQueryValues          url.Values
	HasPatientCharacteristics   bool
	ConditionQueryValues        url.Values
	HasConditionCharacteristics bool
	EncounterQueryValues        url.Values
	HasEncounterCharacteristics bool
}

func LoadCharacteristicInfo(characteristics []fhir.GroupCharacteristicComponent) (*CharacteristicInfo, error) {
	c := new(CharacteristicInfo)
	c.PatientQueryValues = make(url.Values)
	c.ConditionQueryValues = make(url.Values)
	c.EncounterQueryValues = make(url.Values)

	for _, characteristic := range characteristics {
		switch {
		// Age
		case characteristic.Code.MatchesCode("http://loinc.org", "21612-7"):
			lowBD := time.Now().AddDate(-1*int(*characteristic.ValueRange.High.Value)-1, 0, 0)
			lowBDExp := "gt" + lowBD.Format("2006-01-02")
			highBD := time.Now().AddDate(-1*int(*characteristic.ValueRange.Low.Value), 0, 0)
			highBDExp := "lte" + highBD.Format("2006-01-02")
			c.PatientQueryValues.Add("birthdate", lowBDExp)
			c.PatientQueryValues.Add("birthdate", highBDExp)
			c.HasPatientCharacteristics = true
			c.ConditionQueryValues.Add("patient.birthdate", lowBDExp)
			c.ConditionQueryValues.Add("patient.birthdate", highBDExp)
			c.EncounterQueryValues.Add("patient.birthdate", lowBDExp)
			c.EncounterQueryValues.Add("patient.birthdate", highBDExp)

		// Gender
		case characteristic.Code.MatchesCode("http://loinc.org", "21840-4"):
			gender := characteristic.ValueCodeableConcept.Coding[0].Code
			c.PatientQueryValues.Add("gender", gender)
			c.HasPatientCharacteristics = true
			c.ConditionQueryValues.Add("patient.gender", gender)
			c.EncounterQueryValues.Add("patient.gender", gender)

		// Condition
		case characteristic.Code.MatchesCode("http://loinc.org", "11450-4"):
			c.ConditionQueryValues.Add("code", codeableConceptQueryValues(characteristic.ValueCodeableConcept))
			c.HasConditionCharacteristics = true

		// Encounter
		case characteristic.Code.MatchesCode("http://loinc.org", "46240-8"):
			c.EncounterQueryValues.Add("type", codeableConceptQueryValues(characteristic.ValueCodeableConcept))
			c.HasEncounterCharacteristics = true

		// Unknown
		default:
			return nil, fmt.Errorf("Unknown characteristic: %v", characteristic.Code)
		}
	}
	return c, nil
}

func codeableConceptQueryValues(cc *fhir.CodeableConcept) string {
	values := make([]string, len(cc.Coding))
	for i := range cc.Coding {
		values[i] = cc.Coding[i].System + "|" + cc.Coding[i].Code
	}
	return strings.Join(values, ",")
}
