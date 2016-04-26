package groups

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/testutil"
	"github.com/stretchr/testify/suite"
)

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestInstacountSuite(t *testing.T) {
	suite.Run(t, new(InstacountSuite))
}

type InstacountSuite struct {
	testutil.MongoSuite
}

func (suite *InstacountSuite) SetupTest() {
	require := suite.Require()
	assert := suite.Assert()

	// Setup the database
	server.Database = suite.DB()

	// Store the bundle
	bundleFile, err := os.Open("../fixtures/sample-group-data-bundle.json")
	require.NoError(err)

	ctx, rw, _ := gin.CreateTestContext()
	ctx.Request, err = http.NewRequest("POST", "http://ie-server/", bundleFile)
	require.NoError(err)
	ctx.Request.Header.Add("Content-Type", "application/json")
	server.NewBatchController(server.NewMongoDataAccessLayer(suite.DB())).Post(ctx)
	assert.Equal(200, rw.Code)
}

func (suite *InstacountSuite) TearDownTest() {
	suite.TearDownDB()
}

func (suite *InstacountSuite) TearDownSuite() {
	suite.TearDownDBServer()
}

func (suite *InstacountSuite) TestInstaCountAllHandler() {
	require := suite.Require()
	assert := suite.Assert()

	handler := InstaCountAllHandler
	groupFile, _ := os.Open("../fixtures/sample-group.json")

	ctx, w, _ := gin.CreateTestContext()
	ctx.Request, _ = http.NewRequest("POST", "/InstaCountAll", groupFile)
	ctx.Request.Header.Add("Content-Type", "application/json")
	handler(ctx)
	require.Equal(http.StatusOK, w.Code)

	counts := make(map[string]int)
	err := json.NewDecoder(w.Body).Decode(&counts)
	require.NoError(err)

	//TODO: These tests should be made more robust once we have better fixtures and test helpers
	assert.Equal(1, counts["patients"])
	assert.Equal(1, counts["conditions"])
	assert.Equal(1, counts["encounters"])
}

func (suite *InstacountSuite) TestInstaCountAllHandlerWithRefutedCondition() {
	require := suite.Require()
	assert := suite.Assert()

	handler := InstaCountAllHandler
	groupFile, _ := os.Open("../fixtures/sample-group-afib.json")
	ctx, w, _ := gin.CreateTestContext()
	ctx.Request, _ = http.NewRequest("POST", "/InstaCountAll", groupFile)
	ctx.Request.Header.Add("Content-Type", "application/json")
	handler(ctx)
	require.Equal(http.StatusOK, w.Code)

	counts := make(map[string]int)
	err := json.NewDecoder(w.Body).Decode(&counts)

	require.NoError(err)

	//TODO: These tests should be made more robust once we have better fixtures and test helpers
	assert.Equal(0, counts["patients"])
	assert.Equal(0, counts["conditions"])
	assert.Equal(0, counts["encounters"])
}
