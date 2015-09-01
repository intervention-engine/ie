package subscription

import (
	"net/http"
	"net/http/httptest"
	"sync"

	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/server"
	. "gopkg.in/check.v1"
	"gopkg.in/mgo.v2/dbtest"
)

type NotifierSuite struct {
	DBServer          *dbtest.DBServer
	Server            *httptest.Server
	PatientRecieved   string
	TimestampRecieved string
	EndpointRecieved  string
	WorkerChannel     chan ResourceUpdateMessage
}

var _ = Suite(&NotifierSuite{})

func (r *NotifierSuite) SetUpSuite(c *C) {
	//set up dbtest server
	r.DBServer = &dbtest.DBServer{}
	r.DBServer.SetPath(c.MkDir())

	//create test risk server
	r.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.PatientRecieved = req.FormValue("patientId")
		r.TimestampRecieved = req.FormValue("timestamp")
		r.EndpointRecieved = req.FormValue("fhirEndpointUrl")
	}))

	r.WorkerChannel = make(chan ResourceUpdateMessage, 1)
}

func (r *NotifierSuite) SetUpTest(c *C) {
	session := r.DBServer.Session()
	server.Database = session.DB("ie-test")
}

func (r *NotifierSuite) TearDownTest(c *C) {
	server.Database.Session.Close()
	r.DBServer.Wipe()
}

func (r *NotifierSuite) TearDownSuite(c *C) {
	r.Server.Close()
}

func (r *NotifierSuite) TestRiskServiceHandler(c *C) {
	sub := &models.Subscription{}
	channel := &models.SubscriptionChannelComponent{Endpoint: r.Server.URL, Type: "rest-hook"}
	sub.Channel = channel
	server.Database.C("subscriptions").Insert(sub)
	rum := NewResourceUpdateMessage("55c3847267803d2945000003", "2015-04-01T00:00:00-04:00")
	r.WorkerChannel <- rum
	var wg sync.WaitGroup
	wg.Add(1)
	go NotifySubscribers(r.WorkerChannel, "http://example.org", &wg)
	close(r.WorkerChannel)
	wg.Wait()
	c.Assert(r.PatientRecieved, Equals, "55c3847267803d2945000003")
	c.Assert(r.TimestampRecieved, Equals, "2015-04-01T00:00:00-04:00")
	c.Assert(r.EndpointRecieved, Equals, "http://example.org")
}
