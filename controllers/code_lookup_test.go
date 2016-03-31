package controllers

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/utilities"
	"github.com/pebbe/util"
	. "gopkg.in/check.v1"
	"gopkg.in/mgo.v2/dbtest"
)

type CodeLookupSuite struct {
	DBServer *dbtest.DBServer
}

var _ = Suite(&CodeLookupSuite{})

func (l *CodeLookupSuite) SetUpSuite(c *C) {
	l.DBServer = &dbtest.DBServer{}
	l.DBServer.SetPath(c.MkDir())

	file, err := os.Open("../fixtures/code-lookup.json")
	util.CheckErr(err)
	defer file.Close()

	server.Database = l.DBServer.Session().DB("ie-test")
	lookupCollection := server.Database.C("codelookup")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		decoder := json.NewDecoder(strings.NewReader(scanner.Text()))
		entry := &utilities.CodeEntry{}
		err = decoder.Decode(entry)
		util.CheckErr(err)
		lookupCollection.Insert(entry)
	}
	util.CheckErr(scanner.Err())
}

func (l *CodeLookupSuite) TearDownSuite(c *C) {
	server.Database.Session.Close()
	l.DBServer.Stop()
}

func (l *CodeLookupSuite) TestCodeLookupByName(c *C) {
	ctx, w, _ := gin.CreateTestContext()
	namelookupFile, _ := os.Open("../fixtures/sample-lookup-request-by-name.json")
	ctx.Request, _ = http.NewRequest("POST", "/CodeLookup", namelookupFile)
	ctx.Request.Header.Add("Content-Type", "application/json")

	CodeLookup(ctx)
	if w.Code != http.StatusOK {
		c.Fatal("Non-OK response code received: ", w.Code)
	}

	nameResponseCodes := []utilities.CodeEntry{}
	err := json.NewDecoder(w.Body).Decode(&nameResponseCodes)
	util.CheckErr(err)

	c.Assert(len(nameResponseCodes), Equals, 10)
}

func (l *CodeLookupSuite) TestCodeLookupByCode(c *C) {
	ctx, w, _ := gin.CreateTestContext()
	codeLookupFile, _ := os.Open("../fixtures/sample-lookup-request-by-code.json")
	ctx.Request, _ = http.NewRequest("POST", "/CodeLookup", codeLookupFile)
	ctx.Request.Header.Add("Content-Type", "application/json")

	CodeLookup(ctx)
	if w.Code != http.StatusOK {
		c.Fatal("Non-OK response code received: ", w.Code)
	}

	codeResponseCodes := []utilities.CodeEntry{}
	err := json.NewDecoder(w.Body).Decode(&codeResponseCodes)
	util.CheckErr(err)

	c.Assert(len(codeResponseCodes), Equals, 6)
}
