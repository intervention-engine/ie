package models

import (
	"encoding/json"
	"os"

	"github.com/pebbe/util"
	. "gopkg.in/check.v1"
	"gopkg.in/mgo.v2"
)

type FilterSuite struct {
	Session *mgo.Session
}

var _ = Suite(&FilterSuite{})

func (f *FilterSuite) SetUpSuite(c *C) {
	var err error
	// Setup the database
	f.Session, err = mgo.Dial("localhost")
	util.CheckErr(err)
	queryCollection := f.Session.DB("ie-test").C("querys")
	queryCollection.DropCollection()
}

func (f *FilterSuite) TearDownSuite(c *C) {
	f.Session.Close()
}

func (f *FilterSuite) TestCreateQuery(c *C) {
	filter := LoadFilterFromFixture("../fixtures/gender-filter.json")
	queryId, err := filter.CreateQuery(f.Session.DB("ie-test"))
	c.Assert(err, IsNil)
	c.Assert(queryId, NotNil)
	queryCount, _ := f.Session.DB("ie-test").C("querys").Count()
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
