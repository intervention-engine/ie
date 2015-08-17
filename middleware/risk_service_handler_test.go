package middleware

import (
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/intervention-engine/fhir/server"
	"github.com/pebbe/util"
	. "gopkg.in/check.v1"
	"gopkg.in/mgo.v2/dbtest"
	"net/http"
	"net/http/httptest"
	"os"
)

type RiskServiceHandlerSuite struct {
	DBServer          *dbtest.DBServer
	Server            *httptest.Server
	RiskServiceServer *httptest.Server
	PatientRecieved   string
	TimestampRecieved string
	EndpointRecieved  string
}

var _ = Suite(&RiskServiceHandlerSuite{})

func (r *RiskServiceHandlerSuite) SetUpSuite(c *C) {
	//set up dbtest server
	r.DBServer = &dbtest.DBServer{}
	r.DBServer.SetPath(c.MkDir())

	//set up empty router
	router := mux.NewRouter()
	router.StrictSlash(true)
	router.KeepContext = true

	//create test server
	r.Server = httptest.NewUnstartedServer(router)

	//create test risk server
	r.RiskServiceServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.PatientRecieved = req.FormValue("patientId")
		r.TimestampRecieved = req.FormValue("timestamp")
		r.EndpointRecieved = req.FormValue("fhirEndpointUrl")
	}))

	r.Server.Start()
	//create and add middleware config to test server
	mwConfig := map[string][]negroni.Handler{
		"MedicationStatementCreate": []negroni.Handler{GenerateRiskHandler(r.RiskServiceServer.URL, r.Server.URL)}}
	server.RegisterRoutes(router, mwConfig)
}

func (r *RiskServiceHandlerSuite) SetUpTest(c *C) {
	session := r.DBServer.Session()
	server.Database = session.DB("ie-test")
}

func (r *RiskServiceHandlerSuite) TearDownTest(c *C) {
	server.Database.Session.Close()
	r.DBServer.Wipe()
}

func (r *RiskServiceHandlerSuite) TearDownSuite(c *C) {
	r.Server.Close()
	r.RiskServiceServer.Close()
}

func (r *RiskServiceHandlerSuite) TestRiskServiceHandler(c *C) {
	//load fixture
	data, err := os.Open("../fixtures/medication-statement.json")
	defer data.Close()
	util.CheckErr(err)

	//post fixture
	client := &http.Client{}
	req, err := http.NewRequest("POST", r.Server.URL+"/MedicationStatement", data)
	util.CheckErr(err)
	_, err = client.Do(req)
	c.Assert(r.PatientRecieved, Equals, "55c3847267803d2945000003")
	c.Assert(r.TimestampRecieved, Equals, "2015-04-01T00:00:00-04:00")
	c.Assert(r.EndpointRecieved, Equals, r.Server.URL)
}
