package controllers

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/server"
	"github.com/pebbe/util"
	. "gopkg.in/check.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/dbtest"
)

type NotificationCountSuite struct {
	DBServer               *dbtest.DBServer
	NotificationCollection *mgo.Collection
}

func Test(t *testing.T) { TestingT(t) }

var _ = Suite(&NotificationCountSuite{})

func (n *NotificationCountSuite) SetUpSuite(c *C) {
	n.DBServer = &dbtest.DBServer{}
	n.DBServer.SetPath(c.MkDir())
}

func (n *NotificationCountSuite) SetUpTest(c *C) {
	server.Database = n.DBServer.Session().DB("ie-test")
	n.NotificationCollection = server.Database.C("communicationrequests")
}

func (n *NotificationCountSuite) TearDownTest(c *C) {
	server.Database.Session.Close()
	n.DBServer.Wipe()
}

func (n *NotificationCountSuite) TearDownSuite(c *C) {
	n.DBServer.Stop()
}

func (n *NotificationCountSuite) TestEmptyNotificationCount(c *C) {
	handler := NotificationCountHandler
	ctx, w, _ := gin.CreateTestContext()
	ctx.Request, _ = http.NewRequest("GET", "/NotificationCount", nil)
	handler(ctx)
	if w.Code != http.StatusOK {
		c.Fatalf("Non-OK response code received: %v", w.Code)
	}

	var counts []NotificationCountResult
	err := json.NewDecoder(w.Body).Decode(&counts)
	util.CheckErr(err)

	c.Assert(counts, HasLen, 0)
}

func (n *NotificationCountSuite) TestNotificationCount(c *C) {
	notification, err := UnmarshallCommunicationRequest("../fixtures/communication-request.json")
	util.CheckErr(err)

	for i, id := range []string{"a", "b", "c", "b", "b", "a"} {
		notification.Id = string(i)
		notification.Subject.Reference = "http://test-ie/Patient/" + id
		notification.Subject.ReferencedID = id
		err = n.NotificationCollection.Insert(*notification)
		util.CheckErr(err)
	}

	handler := NotificationCountHandler
	ctx, w, _ := gin.CreateTestContext()
	ctx.Request, _ = http.NewRequest("GET", "/NotificationCount", nil)
	handler(ctx)
	if w.Code != http.StatusOK {
		c.Fatalf("Non-OK response code received: %v", w.Code)
	}

	var counts []NotificationCountResult
	err = json.NewDecoder(w.Body).Decode(&counts)
	util.CheckErr(err)

	c.Assert(counts, HasLen, 3)
	m := make(map[string]int)
	for _, count := range counts {
		m[count.Patient] = int(count.Count)
	}
	c.Assert(m["a"], Equals, 2)
	c.Assert(m["b"], Equals, 3)
	c.Assert(m["c"], Equals, 1)
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
