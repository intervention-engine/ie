package middleware

import (
	"bufio"
	"encoding/json"
	"github.com/codegangsta/negroni"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/models"
	"github.com/pebbe/util"
	. "gopkg.in/check.v1"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2/dbtest"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
)

type RiskQueryHandlerSuite struct {
	DBServer *dbtest.DBServer
}

var _ = Suite(&RiskQueryHandlerSuite{})

func (r *RiskQueryHandlerSuite) SetUpSuite(c *C) {
	r.DBServer = &dbtest.DBServer{}
	r.DBServer.SetPath(c.MkDir())
}

func (r *RiskQueryHandlerSuite) SetUpTest(c *C) {
	file, err := os.Open("../fixtures/facts.json")
	defer file.Close()
	util.CheckErr(err)

	// Setup the database
	session := r.DBServer.Session()
	defer session.Close()
	factCollection := session.DB("ie-test").C("facts")
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
}

func (r *RiskQueryHandlerSuite) TearDownTest(c *C) {
	r.DBServer.Wipe()
}

func (r *RiskQueryHandlerSuite) TearDownSuite(c *C) {
	r.DBServer.Stop()
}

func (r *RiskQueryHandlerSuite) TestHandleRiskModelParameters(c *C) {
	rawUrl := "http://foo.com/Patient?_query=risk&Condition=count&ConditionWeight=5.0"
	testUrl, _ := url.Parse(rawUrl)
	rmps := handleRiskModelParameters(testUrl.Query())
	c.Assert(len(rmps), Equals, 1)
	c.Assert(rmps[0].Category, Equals, "Condition")
	c.Assert(rmps[0].Method, Equals, "count")
	c.Assert(rmps[0].Weight, Equals, 5.0)
}

func (r *RiskQueryHandlerSuite) TestRiskQueryHandler(c *C) {
	session := r.DBServer.Session()
	defer session.Close()
	server.Database = session.DB("ie-test")
	n := negroni.New()
	n.UseFunc(RiskQueryHandler)
	testServer := httptest.NewServer(n)
	client := http.Client{}
	response, err := client.Get(testServer.URL + "/Patient?_query=risk&Condition=count&ConditionWeight=5.0")
	util.CheckErr(err)
	c.Assert(response, NotNil)
}
