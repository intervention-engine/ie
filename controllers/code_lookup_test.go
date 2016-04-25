package controllers

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/testutil"
	"github.com/intervention-engine/ie/utilities"
	"github.com/stretchr/testify/suite"
)

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCodeLookupSuite(t *testing.T) {
	suite.Run(t, new(CodeLookupSuite))
}

type CodeLookupSuite struct {
	testutil.MongoSuite
}

func (l *CodeLookupSuite) SetupSuite() {
	server.Database = l.DB()

	var codes []utilities.CodeEntry
	l.InsertFixture("codelookup", "../fixtures/code-lookup.json", &codes)
}

func (l *CodeLookupSuite) TearDownSuite() {
	l.TearDownDBServer()
}

func (l *CodeLookupSuite) TestCodeLookupByName() {
	require := l.Require()
	assert := l.Assert()

	ctx, w, _ := gin.CreateTestContext()
	namelookupFile, _ := os.Open("../fixtures/sample-lookup-request-by-name.json")
	ctx.Request, _ = http.NewRequest("POST", "/CodeLookup", namelookupFile)
	ctx.Request.Header.Add("Content-Type", "application/json")

	CodeLookup(ctx)
	require.Equal(http.StatusOK, w.Code)

	nameResponseCodes := []utilities.CodeEntry{}
	err := json.NewDecoder(w.Body).Decode(&nameResponseCodes)
	require.NoError(err)

	assert.Len(nameResponseCodes, 10)
}

func (l *CodeLookupSuite) TestCodeLookupByCode() {
	require := l.Require()
	assert := l.Assert()

	ctx, w, _ := gin.CreateTestContext()
	codeLookupFile, _ := os.Open("../fixtures/sample-lookup-request-by-code.json")
	ctx.Request, _ = http.NewRequest("POST", "/CodeLookup", codeLookupFile)
	ctx.Request.Header.Add("Content-Type", "application/json")

	CodeLookup(ctx)
	require.Equal(http.StatusOK, w.Code)

	codeResponseCodes := []utilities.CodeEntry{}
	err := json.NewDecoder(w.Body).Decode(&codeResponseCodes)
	require.NoError(err)

	assert.Len(codeResponseCodes, 6)
}
