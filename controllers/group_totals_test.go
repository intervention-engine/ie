package controllers

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
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
	server.Database = q.DBServer.Session().DB("ie-test")

	// Store the bundle
	bundleFile, err := os.Open("../fixtures/sample-group-data-bundle.json")
	util.CheckErr(err)

	ctx, rw, _ := gin.CreateTestContext()
	ctx.Request, err = http.NewRequest("POST", "http://ie-server/", bundleFile)
	util.CheckErr(err)
	ctx.Request.Header.Add("Content-Type", "application/json")
	server.NewBatchController(server.NewMongoDataAccessLayer(server.Database)).Post(ctx)
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

	ctx, w, _ := gin.CreateTestContext()
	ctx.Request, _ = http.NewRequest("POST", "/InstaCountAll", groupFile)
	ctx.Request.Header.Add("Content-Type", "application/json")
	handler(ctx)
	if w.Code != http.StatusOK {
		c.Fatalf("Non-OK response code received: %v", w.Code)
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
	ctx, w, _ := gin.CreateTestContext()
	ctx.Request, _ = http.NewRequest("POST", "/InstaCountAll", groupFile)
	ctx.Request.Header.Add("Content-Type", "application/json")
	handler(ctx)
	if w.Code != http.StatusOK {
		c.Fatalf("Non-OK response code received: %v", w.Code)
	}

	counts := make(map[string]int)
	err := json.NewDecoder(w.Body).Decode(&counts)

	util.CheckErr(err)

	//TODO: These tests should be made more robust once we have better fixtures and test helpers
	c.Assert(counts["patients"], Equals, 0)
	c.Assert(counts["conditions"], Equals, 0)
	c.Assert(counts["encounters"], Equals, 0)
}
