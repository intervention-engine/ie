package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/notifications"
	"github.com/intervention-engine/ie/testutil"
	"github.com/stretchr/testify/suite"
	"gopkg.in/mgo.v2"
)

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestNotificationHandlerSuite(t *testing.T) {
	suite.Run(t, new(NotificationHandlerSuite))
}

type NotificationHandlerSuite struct {
	testutil.MongoSuite
	Server                 *httptest.Server
	Handler                *NotificationHandler
	NotificationCollection *mgo.Collection
}

func (n *NotificationHandlerSuite) SetupTest() {
	server.Database = n.DB()
	n.NotificationCollection = n.DB().C("communicationrequests")

	//register notification handler middleware
	n.Handler = &NotificationHandler{Registry: &notifications.NotificationDefinitionRegistry{}}
	mwConfig := map[string][]gin.HandlerFunc{
		"Encounter": []gin.HandlerFunc{n.Handler.Handle()}}

	//set up routes and middleware
	e := gin.New()

	server.RegisterRoutes(e, mwConfig, server.NewMongoDataAccessLayer(server.Database), server.Config{})

	//create test server
	n.Server = httptest.NewServer(e)
}

func (n *NotificationHandlerSuite) TearDownTest() {
	//clear the notification definition registry and the notification database
	n.Handler.Registry = &notifications.NotificationDefinitionRegistry{}
	n.TearDownDB()
	n.Server.Close()
}

func (n *NotificationHandlerSuite) TearDownSuite() {
	n.TearDownDBServer()
}

func (n *NotificationHandlerSuite) TestNotificationTriggers() {
	require := n.Require()
	assert := n.Assert()

	n.Handler.Registry.Register(new(PlannedEncounterNotificationDefinition))
	//load fixture
	data, err := os.Open("../fixtures/encounter-planned.json")
	require.NoError(err)
	defer data.Close()

	//post fixture
	client := &http.Client{}
	req, err := http.NewRequest("POST", n.Server.URL+"/Encounter", data)
	req.Header.Add("Content-Type", "application/json")
	require.NoError(err)
	res, err := client.Do(req)
	require.NoError(err)
	require.Equal(201, res.StatusCode)

	//check for notification created
	query := n.NotificationCollection.Find(nil)

	//make sure there is only one
	count, err := query.Count()
	require.NoError(err)
	require.Equal(1, count)

	//make sure it is the right one
	result := models.CommunicationRequest{}
	query.One(&result)
	assert.Equal("123", result.Id)
	assert.Equal("http://intervention-engine.org/Patient/5540f2041cd4623133000001", result.Subject.Reference)
}

func (n *NotificationHandlerSuite) TestNotificationDoesNotTrigger() {
	require := n.Require()
	assert := n.Assert()

	n.Handler.Registry.Register(new(PlannedEncounterNotificationDefinition))

	//load fixture
	data, err := os.Open("../fixtures/encounter-office-visit.json")
	require.NoError(err)
	defer data.Close()

	//post fixture
	client := &http.Client{}
	req, err := http.NewRequest("POST", n.Server.URL+"/Encounter", data)
	require.NoError(err)
	req.Header.Add("Content-Type", "application/json")
	_, err = client.Do(req)
	require.NoError(err)

	//check for no notification
	count, err := n.NotificationCollection.Count()
	require.NoError(err)
	assert.Equal(0, count)
}

//  Dummy notification definition for testing
type PlannedEncounterNotificationDefinition struct{}

func (def *PlannedEncounterNotificationDefinition) Name() string {
	return "Planned Encounter"
}
func (def *PlannedEncounterNotificationDefinition) Triggers(resource interface{}, action string) bool {
	enc, ok := resource.(*models.Encounter)
	return action == "create" && ok && enc.Status == "planned"
}
func (def *PlannedEncounterNotificationDefinition) GetNotification(resource interface{}, action string, baseURL string) *models.CommunicationRequest {
	if def.Triggers(resource, action) {
		enc := resource.(*models.Encounter)
		cr := &models.CommunicationRequest{}
		cr.Id = "123"
		cr.Subject = enc.Patient
		return cr
	}
	return nil
}
