package subscription

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/intervention-engine/fhir/server"
	"github.com/pebbe/util"
	. "gopkg.in/check.v1"
	"gopkg.in/mgo.v2/dbtest"
)

type ResourceWatchSuite struct {
	DBServer      *dbtest.DBServer
	Server        *httptest.Server
	WorkerChannel chan ResourceUpdateMessage
}

var _ = Suite(&ResourceWatchSuite{})

func Test(t *testing.T) { TestingT(t) }

func (r *ResourceWatchSuite) SetUpSuite(c *C) {
	//set up dbtest server
	r.DBServer = &dbtest.DBServer{}
	r.DBServer.SetPath(c.MkDir())
}

func (r *ResourceWatchSuite) SetUpTest(c *C) {
	server.Database = r.DBServer.Session().DB("ie-test")

	//set up empty router
	e := gin.New()

	r.WorkerChannel = make(chan ResourceUpdateMessage, 1)
	//create and add middleware config to test server
	mwConfig := map[string][]gin.HandlerFunc{
		"MedicationStatement": []gin.HandlerFunc{GenerateResourceWatch(r.WorkerChannel)}}
	server.RegisterRoutes(e, mwConfig, server.NewMongoDataAccessLayer(server.Database), server.Config{})
	//create test server
	r.Server = httptest.NewUnstartedServer(e)
	r.Server.Start()
}

func (r *ResourceWatchSuite) TearDownTest(c *C) {
	close(r.WorkerChannel)
	server.Database.Session.Close()
	r.DBServer.Wipe()
	r.Server.Close()
}

func (r *ResourceWatchSuite) TearDownSuite(c *C) {
	r.DBServer.Stop()
}

func (r *ResourceWatchSuite) TestGenerateResourceWatch(c *C) {
	//load fixture
	data, err := os.Open("../fixtures/medication-statement.json")
	util.CheckErr(err)
	defer data.Close()

	//post fixture
	client := &http.Client{}
	req, err := http.NewRequest("POST", r.Server.URL+"/MedicationStatement", data)
	util.CheckErr(err)
	req.Header.Add("Content-Type", "application/json")
	_, err = client.Do(req)

	util.CheckErr(err)
	c.Assert(len(r.WorkerChannel), Equals, 1)
	rum := <-r.WorkerChannel
	c.Assert(rum.PatientID, Equals, "55c3847267803d2945000003")
	c.Assert(rum.Timestamp, Equals, "2015-04-01T00:00:00-04:00")
}
