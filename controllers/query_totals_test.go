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
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

type QueryTotalsSuite struct {
	Session *mgo.Session
}

var _ = Suite(&QueryTotalsSuite{})

func (q *QueryTotalsSuite) SetUpSuite(c *C) {
	file, err := os.Open("../fixtures/facts.json")
	defer file.Close()
	util.CheckErr(err)

	// Setup the database
	q.Session, err = mgo.Dial("localhost")
	util.CheckErr(err)
	factCollection := q.Session.DB("ie-test").C("facts")
	factCollection.DropCollection()
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
	server.Database = q.Session.DB("ie-test")
}

func (q *QueryTotalsSuite) TearDownSuite(c *C) {
	q.Session.Close()
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
