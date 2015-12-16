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
	name_req, _ := http.NewRequest("POST", "/CodeLookup", namelookupFile)
	w := httptest.NewRecorder()
	handler(w, name_req)
	if w.Code != http.StatusOK {
		c.Fatal("Non-OK response code received: ", w.Code)
	}

	name_response_codes := []utilities.CodeEntry{}
	err := json.NewDecoder(w.Body).Decode(&name_response_codes)
	util.CheckErr(err)

	c.Assert(len(name_response_codes), Equals, 10)
}

func (l *CodeLookupSuite) TestCodeLookupByCode(c *C) {
	handler := CodeLookup
	codelookupFile, _ := os.Open("../fixtures/sample-lookup-request-by-code.json")
	code_req, _ := http.NewRequest("POST", "/CodeLookup", codelookupFile)
	w := httptest.NewRecorder()
	handler(w, code_req)
	if w.Code != http.StatusOK {
		c.Fatal("Non-OK response code received: ", w.Code)
	}

	code_response_codes := []utilities.CodeEntry{}
	err := json.NewDecoder(w.Body).Decode(&code_response_codes)
	util.CheckErr(err)

	c.Assert(len(code_response_codes), Equals, 6)
}
