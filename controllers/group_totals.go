package controllers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	fhirmodels "github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/search"
	"github.com/intervention-engine/fhir/server"
	"github.com/labstack/echo"
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
	//NOTE: We can't prefilter unconfirmed conditions (FHIR doesn't have a search parameter for
	//verificationStatus) and our API doesn't let us insert non-FHIR criteria, so we must
	// post-process the results to filter them out.
	if hasConditionCriteria {
		var cResults []struct {
			Patient patientContainer `bson:"patient"`
			Status  string           `bson:"verificationStatus"`
		}
		cSearchQuery := search.Query{Resource: "Condition", Query: cquery}
		cQ := searcher.CreateQueryWithoutOptions(cSearchQuery)

		cQ.Select(bson.M{"patient.referenceid": 1, "verificationStatus": 1}).All(&cResults)
		cids = make([]string, 0, len(cResults))
		for i := range cResults {
			if cResults[i].Status == "confirmed" {
				cids = append(cids, cResults[i].Patient.ID)
			}
		}
	}

	//We only need to query for encounters if the group contains an encounter criteria
	if hasEncounterCriteria {
		var eResultIDs []struct {
			Patient patientContainer `bson:"patient"`
		}
		eSearchQuery := search.Query{Resource: "Encounter", Query: equery}
		eQ := searcher.CreateQueryWithoutOptions(eSearchQuery)
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
		pQ := searcher.CreateQueryWithoutOptions(pSearchQuery)
		pQ.Select(bson.M{"_id": 1}).All(&pResultIDs)
		pids = make([]string, len(pResultIDs))
		for i := range pResultIDs {
			pids[i] = pResultIDs[i].ID
		}
	}

	return pids
}

func PatientListHandler(c *echo.Context) error {
	groupController := server.NewResourceController("Group", server.NewMongoDataAccessLayer(server.Database))
	group, err := groupController.LoadResource(c)
	if err != nil {
		return err
	}
	patientIds := groupListResolver(*group.(*fhirmodels.Group))
	responseMap := map[string][]string{
		"patientids": patientIds,
	}
	return c.JSON(http.StatusOK, responseMap)
}

// PatientListMiddleware is a tremendous hack just to get a custom _query=group working quickly so the UI team
// can use it.  Aside from being ugly and non-general, one major shortcoming is that the navigation links that are
// returned have the expanded id set rather than the _query and groupId.
func PatientListMiddleware() echo.MiddlewareFunc {
	return func(hf echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if c.Request().Method == "GET" && c.Request().URL.Query().Get("_query") == "group" {
				qv := c.Request().URL.Query()

				// First get the group
				var groupId bson.ObjectId
				if bson.IsObjectIdHex(qv.Get("groupId")) {
					groupId = bson.ObjectIdHex(qv.Get("groupId"))
				} else {
					return errors.New("Invalid id")
				}
				var group fhirmodels.Group
				err := server.Database.C("groups").Find(bson.M{"_id": groupId.Hex()}).One(&group)
				if err != nil {
					return err
				}
				patientIds := groupListResolver(group)

				// Since we're going to pass the context on to another controller, fixup the query, substituting
				// the _query and groupId with an expanded list of ids
				qv.Del("_query")
				qv.Del("groupId")
				qv.Add("_id", strings.Join(patientIds, ","))
				c.Request().URL.RawQuery = qv.Encode()

				// Pass it on to the patient controller
				patientController := server.NewResourceController("Patient", server.NewMongoDataAccessLayer(server.Database))
				return patientController.IndexHandler(c)
			}

			return hf(c)
		}
	}
}

func trimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}

func InstaCountAllHandler(c *echo.Context) error {
	group := &fhirmodels.Group{}
	err := c.Bind(group)
	if err != nil {
		return err
	}

	pids := groupListResolver(*group)
	pCount := len(pids)

	cCollection := server.Database.C("conditions")
	// Only count the confirmed conditions
	cCriteria := bson.M{
		"patient.referenceid": bson.M{"$in": pids},
		"verificationStatus":  "confirmed",
	}
	cCount, err := cCollection.Find(cCriteria).Count()

	eCollection := server.Database.C("encounters")
	eCount, err := eCollection.Find(bson.M{"patient.referenceid": bson.M{"$in": pids}}).Count()

	newResultMap := map[string]int{
		"patients":   pCount,
		"conditions": cCount,
		"encounters": eCount,
	}

	return c.JSON(http.StatusOK, newResultMap)
}
