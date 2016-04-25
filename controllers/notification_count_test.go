package controllers

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/testutil"
	"github.com/stretchr/testify/suite"
	"gopkg.in/mgo.v2"
)

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestNotificationCountSuite(t *testing.T) {
	suite.Run(t, new(NotificationCountSuite))
}

type NotificationCountSuite struct {
	testutil.MongoSuite
	NotificationCollection *mgo.Collection
}

func (n *NotificationCountSuite) SetupTest() {
	server.Database = n.DB()
	n.NotificationCollection = server.Database.C("communicationrequests")
}

func (n *NotificationCountSuite) TearDownTest() {
	n.TearDownDB()
}

func (n *NotificationCountSuite) TearDownSuite() {
	n.TearDownDBServer()
}

func (n *NotificationCountSuite) TestEmptyNotificationCount() {
	require := n.Require()
	assert := n.Assert()

	handler := NotificationCountHandler
	ctx, w, _ := gin.CreateTestContext()
	ctx.Request, _ = http.NewRequest("GET", "/NotificationCount", nil)
	handler(ctx)
	require.Equal(http.StatusOK, w.Code)

	var counts []NotificationCountResult
	err := json.NewDecoder(w.Body).Decode(&counts)
	require.NoError(err)

	assert.Empty(counts)
}

func (n *NotificationCountSuite) TestNotificationCount() {
	require := n.Require()
	assert := n.Assert()

	notification, err := UnmarshallCommunicationRequest("../fixtures/communication-request.json")
	require.NoError(err)

	for i, id := range []string{"a", "b", "c", "b", "b", "a"} {
		notification.Id = string(i)
		notification.Subject.Reference = "http://test-ie/Patient/" + id
		notification.Subject.ReferencedID = id
		err = n.NotificationCollection.Insert(*notification)
		require.NoError(err)
	}

	handler := NotificationCountHandler
	ctx, w, _ := gin.CreateTestContext()
	ctx.Request, _ = http.NewRequest("GET", "/NotificationCount", nil)
	handler(ctx)
	require.Equal(http.StatusOK, w.Code)

	var counts []NotificationCountResult
	err = json.NewDecoder(w.Body).Decode(&counts)
	require.NoError(err)

	assert.Len(counts, 3)
	m := make(map[string]int)
	for _, count := range counts {
		m[count.Patient] = int(count.Count)
	}
	assert.Equal(2, m["a"])
	assert.Equal(3, m["b"])
	assert.Equal(1, m["c"])
}

func UnmarshallCommunicationRequest(file string) (*models.CommunicationRequest, error) {
	data, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer data.Close()

	cr := &models.CommunicationRequest{}
	err = json.NewDecoder(data).Decode(cr)
	return cr, err
}
