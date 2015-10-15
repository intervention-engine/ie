package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"

	fhirmodels "github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/server"
	"github.com/pebbe/util"
	. "gopkg.in/check.v1"
	"gopkg.in/mgo.v2/dbtest"
)

type QueryTotalsSuite struct {
	DBServer *dbtest.DBServer
}

var _ = Suite(&QueryTotalsSuite{})

func (q *QueryTotalsSuite) SetUpSuite(c *C) {
	q.DBServer = &dbtest.DBServer{}
	q.DBServer.SetPath(c.MkDir())
}

func (q *QueryTotalsSuite) SetUpTest(c *C) {
	patientFile, err := os.Open("../fixtures/patient-example-a.json")
	defer patientFile.Close()
	util.CheckErr(err)

	encounterFile, err := os.Open("../fixtures/encounter-example.json")
	defer encounterFile.Close()
	util.CheckErr(err)

	conditionFile, err := os.Open("../fixtures/condition-example.json")
	defer conditionFile.Close()
	util.CheckErr(err)

	// Setup the database
	session := q.DBServer.Session()

	patientCollection := session.DB("ie-test").C("patients")
	encounterCollection := session.DB("ie-test").C("encounters")
	conditionCollection := session.DB("ie-test").C("conditions")

	patientDecoder := json.NewDecoder(patientFile)
	encounterDecoder := json.NewDecoder(encounterFile)
	conditionDecoder := json.NewDecoder(conditionFile)

	patient := &fhirmodels.Patient{}
	encounter := &fhirmodels.Encounter{}
	condition := &fhirmodels.Condition{}

	err = patientDecoder.Decode(patient)
	util.CheckErr(err)
	err = encounterDecoder.Decode(encounter)
	util.CheckErr(err)
	err = conditionDecoder.Decode(condition)
	util.CheckErr(err)

	patient.Id = "TESTID"

	patientCollection.Insert(patient)
	encounterCollection.Insert(encounter)
	conditionCollection.Insert(condition)

	server.Database = session.DB("ie-test")
}

func (q *QueryTotalsSuite) TearDownTest(c *C) {
	server.Database.Session.Close()
	q.DBServer.Wipe()
}

func (q *QueryTotalsSuite) TearDownSuite(c *C) {
	q.DBServer.Stop()
}

func (q *QueryTotalsSuite) TestInstaCountAllHandler(c *C) {
	handler := InstaCountAllHandler
	groupFile, _ := os.Open("../fixtures/sample-group.json")
	req, _ := http.NewRequest("POST", "/InstaCountAll", groupFile)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		c.Fatal("Non-OK response code received: %v", w.Code)
	}

	counts := make(map[string]int)
	err := json.NewDecoder(w.Body).Decode(&counts)

	util.CheckErr(err)

	//TODO: These tests should be made more robust once we have better fixtures and test helpers
	c.Assert(counts["patients"], Equals, 1)
	c.Assert(counts["conditions"], Equals, 1)
	c.Assert(counts["encounters"], Equals, 1)
}
