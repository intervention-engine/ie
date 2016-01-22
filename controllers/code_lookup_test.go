package controllers

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/utilities"
	"github.com/labstack/echo"
	"github.com/pebbe/util"
	. "gopkg.in/check.v1"
	"gopkg.in/mgo.v2/dbtest"
)

type CodeLookupSuite struct {
	DBServer *dbtest.DBServer
	Echo     *echo.Echo
}

var _ = Suite(&CodeLookupSuite{})

func (l *CodeLookupSuite) SetUpSuite(c *C) {
	l.DBServer = &dbtest.DBServer{}
	l.DBServer.SetPath(c.MkDir())
	l.Echo = echo.New()

	file, err := os.Open("../fixtures/code-lookup.json")
	util.CheckErr(err)
	defer file.Close()

	session := l.DBServer.Session()
	lookupCollection := session.DB("ie-test").C("codelookup")
	lookupCollection.DropCollection()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		decoder := json.NewDecoder(strings.NewReader(scanner.Text()))
		entry := &utilities.CodeEntry{}
		err = decoder.Decode(entry)
		util.CheckErr(err)
		lookupCollection.Insert(entry)
	}
	util.CheckErr(scanner.Err())

	server.Database = session.DB("ie-test")
}

func (l *CodeLookupSuite) TearDownSuite(c *C) {
	server.Database.Session.Close()
	l.DBServer.Wipe()
}

func (l *CodeLookupSuite) TestCodeLookupByName(c *C) {
	handler := CodeLookup
	namelookupFile, _ := os.Open("../fixtures/sample-lookup-request-by-name.json")
	nameReq, _ := http.NewRequest("POST", "/CodeLookup", namelookupFile)
	nameReq.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := echo.NewContext(nameReq, echo.NewResponse(w, l.Echo), l.Echo)
	handler(ctx)
	if w.Code != http.StatusOK {
		c.Fatal("Non-OK response code received: ", w.Code)
	}

	nameResponseCodes := []utilities.CodeEntry{}
	err := json.NewDecoder(w.Body).Decode(&nameResponseCodes)
	util.CheckErr(err)

	c.Assert(len(nameResponseCodes), Equals, 10)
}

func (l *CodeLookupSuite) TestCodeLookupByCode(c *C) {
	handler := CodeLookup
	codeLookupFile, _ := os.Open("../fixtures/sample-lookup-request-by-code.json")
	codeReq, _ := http.NewRequest("POST", "/CodeLookup", codeLookupFile)
	w := httptest.NewRecorder()
	codeReq.Header.Add("Content-Type", "application/json")
	ctx := echo.NewContext(codeReq, echo.NewResponse(w, l.Echo), l.Echo)

	err := handler(ctx)
	util.CheckErr(err)
	if w.Code != http.StatusOK {
		c.Fatal("Non-OK response code received: ", w.Code)
	}

	codeResponseCodes := []utilities.CodeEntry{}
	err = json.NewDecoder(w.Body).Decode(&codeResponseCodes)
	util.CheckErr(err)

	c.Assert(len(codeResponseCodes), Equals, 6)
}
