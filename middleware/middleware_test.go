package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	fhirmodels "github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/models"
	"github.com/pebbe/util"
	. "gopkg.in/check.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2/dbtest"
)

type MiddlewareSuite struct {
	DBServer          *dbtest.DBServer
	Server            *httptest.Server
	Router            *mux.Router
	MiddlewareConfig  map[string][]negroni.Handler
	FactCollection    *mgo.Collection
	PatientCollection *mgo.Collection
}

func Test(t *testing.T) { TestingT(t) }

var _ = Suite(&MiddlewareSuite{})

func (m *MiddlewareSuite) SetUpSuite(c *C) {
	m.DBServer = &dbtest.DBServer{}
	m.DBServer.SetPath(c.MkDir())

	//register facthandler middleware
	m.MiddlewareConfig = make(map[string][]negroni.Handler)
	m.MiddlewareConfig["PatientCreate"] = append(m.MiddlewareConfig["PatientCreate"], negroni.HandlerFunc(FactHandler))
	m.MiddlewareConfig["PatientUpdate"] = append(m.MiddlewareConfig["PatientUpdate"], negroni.HandlerFunc(FactHandler))
	m.MiddlewareConfig["PatientDelete"] = append(m.MiddlewareConfig["PatientDelete"], negroni.HandlerFunc(FactHandler))

	//set up routes and middleware
	m.Router = mux.NewRouter()
	m.Router.StrictSlash(true)
	m.Router.KeepContext = true
	server.RegisterRoutes(m.Router, m.MiddlewareConfig)

	//create test server
	m.Server = httptest.NewServer(m.Router)
}

func (m *MiddlewareSuite) SetUpTest(c *C) {
	session := m.DBServer.Session()
	db := session.DB("ie-test")
	server.Database = db
	m.FactCollection = db.C("facts")
	m.PatientCollection = db.C("patients")
}

func (m *MiddlewareSuite) TearDownTest(c *C) {
	server.Database.Session.Close()
	m.DBServer.Wipe()
}

func (m *MiddlewareSuite) TearDownSuite(c *C) {
	m.DBServer.Stop()
	m.Server.Close()
}

func (m *MiddlewareSuite) TestFactCreation(c *C) {
	//load fixture
	data, err := os.Open("../fixtures/patient-example-a.json")
	defer data.Close()
	util.CheckErr(err)

	//post fixture
	client := &http.Client{}
	req, err := http.NewRequest("POST", m.Server.URL+"/Patient", data)
	util.CheckErr(err)
	_, err = client.Do(req)

	//check for fact created
	count, err := m.FactCollection.Count()
	c.Assert(count, Equals, 1)
}

func (m *MiddlewareSuite) TestFactDelete(c *C) {
	//create dummy fact and patient
	i := bson.NewObjectId()
	deleteId := i.Hex()
	m.FactCollection.Insert(models.Fact{Gender: "M", TargetID: deleteId})
	m.PatientCollection.Insert(fhirmodels.Patient{Id: deleteId})

	//create and send http delete request
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", m.Server.URL+"/Patient/"+deleteId, nil)
	util.CheckErr(err)
	_, err = client.Do(req)

	//check that fact is gone
	count, err := m.FactCollection.Find(bson.M{"targetid": deleteId}).Count()
	c.Assert(count, Equals, 0)
}

func (m *MiddlewareSuite) TestFactUpdate(c *C) {
	//load fixture
	data, err := os.Open("../fixtures/patient-example-b.json")
	defer data.Close()
	util.CheckErr(err)

	//create dummy fact and patient
	i := bson.NewObjectId()
	tempId := i.Hex()
	m.FactCollection.Insert(models.Fact{Gender: "M", TargetID: tempId})
	m.PatientCollection.Insert(fhirmodels.Patient{Id: tempId})

	//create and send http put request
	client := &http.Client{}
	req, err := http.NewRequest("PUT", m.Server.URL+"/Patient/"+tempId, data)
	util.CheckErr(err)
	_, err = client.Do(req)

	//check to see that dummy fact was updated from male to female
	fact := models.Fact{}
	err = m.FactCollection.Find(bson.M{"targetid": tempId}).One(&fact)
	util.CheckErr(err)
	c.Assert(fact.Gender, Equals, "female")
}
