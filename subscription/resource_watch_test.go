package subscription

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/testutil"
	"github.com/stretchr/testify/suite"
)

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestResourceWatchSuite(t *testing.T) {
	suite.Run(t, new(ResourceWatchSuite))
}

type ResourceWatchSuite struct {
	testutil.MongoSuite
	Server        *httptest.Server
	WorkerChannel chan ResourceUpdateMessage
}

func (r *ResourceWatchSuite) SetupTest() {
	server.Database = r.DB()

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

func (r *ResourceWatchSuite) TearDownTest() {
	close(r.WorkerChannel)
	r.TearDownDB()
	r.Server.Close()
}

func (r *ResourceWatchSuite) TearDownSuite() {
	r.TearDownDBServer()
}

func (r *ResourceWatchSuite) TestGenerateResourceWatch() {
	require := r.Require()
	assert := r.Assert()

	//load fixture
	data, err := os.Open("../fixtures/medication-statement.json")
	require.NoError(err)
	defer data.Close()

	//post fixture
	client := &http.Client{}
	req, err := http.NewRequest("POST", r.Server.URL+"/MedicationStatement", data)
	require.NoError(err)
	req.Header.Add("Content-Type", "application/json")
	_, err = client.Do(req)

	require.NoError(err)
	assert.Len(r.WorkerChannel, 1)
	rum := <-r.WorkerChannel
	assert.Equal("55c3847267803d2945000003", rum.PatientID)
	assert.Equal("2015-04-01T00:00:00-04:00", rum.Timestamp)
}
