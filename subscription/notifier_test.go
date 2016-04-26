package subscription

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/testutil"
	"github.com/stretchr/testify/suite"
)

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestNotifierSuite(t *testing.T) {
	suite.Run(t, new(NotifierSuite))
}

type NotifierSuite struct {
	testutil.MongoSuite
	Server            *httptest.Server
	PatientRecieved   string
	TimestampRecieved string
	EndpointRecieved  string
	WorkerChannel     chan ResourceUpdateMessage
}

func (r *NotifierSuite) SetupSuite() {
	//create test risk server
	r.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.PatientRecieved = req.FormValue("patientId")
		r.TimestampRecieved = req.FormValue("timestamp")
		r.EndpointRecieved = req.FormValue("fhirEndpointUrl")
	}))

	r.WorkerChannel = make(chan ResourceUpdateMessage, 1)
}

func (r *NotifierSuite) SetupTest() {
	server.Database = r.DB()
}

func (r *NotifierSuite) TearDownTest() {
	r.TearDownDB()
}

func (r *NotifierSuite) TearDownSuite() {
	r.TearDownDBServer()
	r.Server.Close()
}

func (r *NotifierSuite) TestRiskServiceHandler() {
	assert := r.Assert()

	sub := &models.Subscription{}
	channel := &models.SubscriptionChannelComponent{Endpoint: r.Server.URL, Type: "rest-hook"}
	sub.Channel = channel
	r.DB().C("subscriptions").Insert(sub)
	rum := NewResourceUpdateMessage("55c3847267803d2945000003", "2015-04-01T00:00:00-04:00")
	r.WorkerChannel <- rum
	var wg sync.WaitGroup
	wg.Add(1)
	go NotifySubscribers(r.WorkerChannel, "http://example.org", &wg)
	close(r.WorkerChannel)
	wg.Wait()
	assert.Equal("55c3847267803d2945000003", r.PatientRecieved)
	assert.Equal("2015-04-01T00:00:00-04:00", r.TimestampRecieved)
	assert.Equal("http://example.org", r.EndpointRecieved)
}
