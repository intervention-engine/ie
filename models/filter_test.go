package models

import (
	"encoding/json"
	"os"

	"github.com/pebbe/util"
	. "gopkg.in/check.v1"
	"gopkg.in/mgo.v2/dbtest"
)

type FilterSuite struct {
	DBServer *dbtest.DBServer
}

var _ = Suite(&FilterSuite{})

func (f *FilterSuite) SetUpSuite(c *C) {
	f.DBServer = &dbtest.DBServer{}
	f.DBServer.SetPath(c.MkDir())
}

func (f *FilterSuite) TearDownTest(c *C) {
	f.DBServer.Wipe()
}

func (f *FilterSuite) TearDownSuite(c *C) {
	f.DBServer.Stop()
}

func (f *FilterSuite) TestCreateQuery(c *C) {
	filter := LoadFilterFromFixture("../fixtures/gender-filter.json")
	session := f.DBServer.Session()
	defer session.Close()
	queryId, err := filter.CreateQuery(session.DB("ie-test"))
	c.Assert(err, IsNil)
	c.Assert(queryId, NotNil)
	queryCount, _ := session.DB("ie-test").C("querys").Count()
	c.Assert(queryCount, Equals, 1)
}

func LoadFilterFromFixture(fileName string) *Filter {
	data, err := os.Open(fileName)
	defer data.Close()
	util.CheckErr(err)
	decoder := json.NewDecoder(data)
	filter := &Filter{}
	err = decoder.Decode(filter)
	util.CheckErr(err)
	return filter
}
