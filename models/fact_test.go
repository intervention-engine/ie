package models

import (
	"bufio"
	"encoding/json"
	"github.com/pebbe/util"
	"gitlab.mitre.org/intervention-engine/fhir/models"
	. "gopkg.in/check.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"os"
	"strings"
	"testing"
)

type FactSuite struct {
	Session *mgo.Session
	Query   *models.Query
}

type QueryResult struct {
	Total int `json:"total", bson:"total"`
}

func Test(t *testing.T) { TestingT(t) }

var _ = Suite(&FactSuite{})

func (f *FactSuite) SetUpSuite(c *C) {
	file, err := os.Open("../fixtures/facts.json")
	defer file.Close()
	util.CheckErr(err)

	// Setup the database
	f.Session, err = mgo.Dial("localhost")
	util.CheckErr(err)
	factCollection := f.Session.DB("ie-test").C("facts")
	factCollection.DropCollection()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		decoder := json.NewDecoder(strings.NewReader(scanner.Text()))
		fact := &Fact{}
		err = decoder.Decode(fact)
		util.CheckErr(err)
		i := bson.NewObjectId()
		fact.Id = i.Hex()
		factCollection.Insert(fact)
	}
	util.CheckErr(scanner.Err())
	f.Query = LoadQueryFromFixture("../fixtures/sample-query.json")
}

func (f *FactSuite) TearDownSuite(c *C) {
	f.Session.Close()
}

func (f *FactSuite) TestFactFromPatient(c *C) {
	patient := LoadPatientFromFixture("../fixtures/patient-example-a.json")
	fact := FactFromPatient(patient)
	c.Assert(fact.Gender, Equals, "M")
}

func (f *FactSuite) TestCreatePersonPipeline(c *C) {
	pipeline := CreatePersonPipeline(f.Query)
	c.Assert(4, Equals, len(pipeline))
	qr := &QueryResult{}
	factCollection := f.Session.DB("ie-test").C("facts")
	err := factCollection.Pipe(pipeline).One(qr)
	util.CheckErr(err)
	c.Assert(qr.Total, Equals, 5)
}

func (f *FactSuite) TestEmptyQuery(c *C) {
	qr := &QueryResult{}
	factCollection := f.Session.DB("ie-test").C("facts")
	err := factCollection.Pipe(CreatePersonPipeline(&models.Query{})).One(qr)
	util.CheckErr(err)
	c.Assert(qr.Total, Equals, 39)
}

func LoadPatientFromFixture(fileName string) *models.Patient {
	data, err := os.Open(fileName)
	defer data.Close()
	util.CheckErr(err)
	decoder := json.NewDecoder(data)
	patient := &models.Patient{}
	err = decoder.Decode(patient)
	util.CheckErr(err)
	return patient
}

func LoadQueryFromFixture(fileName string) *models.Query {
	data, err := os.Open(fileName)
	defer data.Close()
	util.CheckErr(err)
	decoder := json.NewDecoder(data)
	query := &models.Query{}
	err = decoder.Decode(query)
	util.CheckErr(err)
	return query
}
