package controllers

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/models"
	"github.com/pebbe/util"
	. "gopkg.in/check.v1"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2/dbtest"
	"strings"
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
	file, err := os.Open("../fixtures/facts.json")
	defer file.Close()
	util.CheckErr(err)

	// Setup the database
	session := q.DBServer.Session()
	factCollection := session.DB("ie-test").C("facts")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		decoder := json.NewDecoder(strings.NewReader(scanner.Text()))
		fact := &models.Fact{}
		err = decoder.Decode(fact)
		util.CheckErr(err)
		i := bson.NewObjectId()
		fact.Id = i.Hex()
		factCollection.Insert(fact)
	}
	util.CheckErr(scanner.Err())
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
	queryFile, _ := os.Open("../fixtures/sample-query.json")
	req, _ := http.NewRequest("POST", "/InstaCountAll", queryFile)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		c.Fatal("Non-OK response code received: %v", w.Code)
	}

	counts := make(map[string]int)
	err := json.NewDecoder(w.Body).Decode(&counts)
	util.CheckErr(err)

	c.Assert(counts["patients"], Equals, 5)
	c.Assert(counts["conditions"], Equals, 10)
	c.Assert(counts["encounters"], Equals, 12)
}
