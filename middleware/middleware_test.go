package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/context"
	fhirmodels "github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/models"
	"github.com/pebbe/util"
	. "gopkg.in/check.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MiddlewareSuite struct {
	Session   *mgo.Session
	Server    *httptest.Server
	FixtureId string
}

func Test(t *testing.T) { TestingT(t) }

var _ = Suite(&MiddlewareSuite{})

func (m *MiddlewareSuite) SetUpSuite(c *C) {
	//Set up the database
	var err error
	m.Session, err = mgo.Dial("localhost")
	util.CheckErr(err)
	server.Database = m.Session.DB("ie-test")

	//create test server
	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		patient := &fhirmodels.Patient{}
		if r.Method != "DELETE" {
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(patient)
			util.CheckErr(err)
			i := bson.NewObjectId()
			patient.Id = i.Hex()
		}

		switch r.Method {
		case "POST":
			context.Set(r, "Patient", patient)
			context.Set(r, "Resource", "Patient")
			context.Set(r, "Action", "create")
		case "PUT":
			patient.Id = m.FixtureId
			context.Set(r, "Patient", patient)
			context.Set(r, "Resource", "Patient")
			context.Set(r, "Action", "update")
		case "DELETE":
			context.Set(r, "Patient", r.Body)
			context.Set(r, "Resource", "Patient")
			context.Set(r, "Action", "delete")
		}
		FactHandler(w, r, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	}))
}

func (m *MiddlewareSuite) TearDownSuite(c *C) {
	server.Database.DropDatabase()
	m.Session.Close()
	m.Server.Close()
}

func (m *MiddlewareSuite) TestFactCreation(c *C) {
	data, err := os.Open("../fixtures/patient-example-a.json")
	defer data.Close()
	util.CheckErr(err)

	_, err = http.Post(m.Server.URL, "application/json", data)
	util.CheckErr(err)

	factCollection := server.Database.C("facts")

	count, err := factCollection.Count()
	c.Assert(count, Equals, 1)
}

func (m *MiddlewareSuite) TestFactUpdate(c *C) {
	data, err := os.Open("../fixtures/patient-example-b.json")
	defer data.Close()
	util.CheckErr(err)

	factCollection := server.Database.C("facts")
	i := bson.NewObjectId()
	tempId := i.Hex()
	m.FixtureId = tempId
	factCollection.Insert(models.Fact{Gender:"M", TargetID: tempId})

	client := &http.Client{}
	req, err := http.NewRequest("PUT", m.Server.URL, data)
	util.CheckErr(err)
	_, err = client.Do(req)

	fact := models.Fact{}
	err = factCollection.Find(bson.M{"targetid": tempId}).One(&fact)
	util.CheckErr(err)

	c.Assert(fact.Gender, Equals, "F")
}

// func (m *MiddlewareSuite) TestFactDelete(c *C) {
// 	factCollection := server.Database.C("facts")
// 	i := bson.NewObjectId()
// 	tempId := i.Hex()
// 	factCollection.Insert(models.Fact{Gender:"M", TargetID: tempId})
//
// 	client := &http.Client{}
// 	req, err := http.NewRequest("DELETE", m.Server.URL)
// 	util.CheckErr(err)
// 	_, err = client.Do(req)
//
// 	count, err = factCollection.Find(bson.M{"targetid": tempId}).Count()
// 	c.Assert(count, Equals, 0)
// }
