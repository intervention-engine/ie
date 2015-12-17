package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"

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
	// Setup the database
	session := q.DBServer.Session()
	server.Database = session.DB("ie-test")

	// Store the bundle
	bundleFile, err := os.Open("../fixtures/sample-group-data-bundle.json")
	util.CheckErr(err)
	r, err := http.NewRequest("POST", "http://ie-server/", bundleFile)
	util.CheckErr(err)
	rw := httptest.NewRecorder()
	server.BatchHandler(rw, r, nil)
	c.Assert(rw.Code, Equals, 200)
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

func (q *QueryTotalsSuite) TestInstaCountAllHandlerWithRefutedCondition(c *C) {
	handler := InstaCountAllHandler
	groupFile, _ := os.Open("../fixtures/sample-group-afib.json")
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
	c.Assert(counts["patients"], Equals, 0)
	c.Assert(counts["conditions"], Equals, 0)
	c.Assert(counts["encounters"], Equals, 0)
}
