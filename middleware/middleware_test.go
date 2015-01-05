package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/models"
	"github.com/pebbe/util"
	. "gopkg.in/check.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MiddlewareSuite struct {
	Session          *mgo.Session
	Server           *httptest.Server
	Router           *mux.Router
	MiddlewareConfig map[string][]negroni.Handler
	FixtureId        string
}

func Test(t *testing.T) { TestingT(t) }

var _ = Suite(&MiddlewareSuite{})

func (m *MiddlewareSuite) SetUpSuite(c *C) {
	//Set up the database
	var err error
	m.Session, err = mgo.Dial("localhost")
	util.CheckErr(err)
	server.Database = m.Session.DB("ie-test")
	factCollection := server.Database.C("facts")
	factCollection.DropCollection()

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

func (m *MiddlewareSuite) TearDownSuite(c *C) {
	server.Database.C("facts").DropCollection()
	m.Session.Close()
	m.Server.Close()
}

func (m *MiddlewareSuite) TestFactCreation(c *C) {
	data, err := os.Open("../fixtures/patient-example-a.json")
	defer data.Close()
	util.CheckErr(err)

	_, err = http.Post(m.Server.URL+"/Patient", "application/json", data)
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
	factCollection.Insert(models.Fact{Gender: "M", TargetID: tempId})

	client := &http.Client{}
	req, err := http.NewRequest("PUT", m.Server.URL+"/Patient/"+tempId, data)
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
// 	deleteId := i.Hex()
// 	factCollection.Insert(models.Fact{Gender: "M", TargetID: deleteId})
//
// 	client := &http.Client{}
// 	req, err := http.NewRequest("DELETE", m.Server.URL+"/Patient/"+deleteId, nil)
// 	util.CheckErr(err)
// 	_, err = client.Do(req)
//
// 	fmt.Println("delete id after req:", deleteId)
// 	count, err := factCollection.Find(bson.M{"targetid": deleteId}).Count()
// 	c.Assert(count, Equals, 0)
// }
