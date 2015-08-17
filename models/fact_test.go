package models

import (
	"encoding/json"
	"github.com/intervention-engine/fhir/models"
	"github.com/pebbe/util"
	. "gopkg.in/check.v1"
	"os"
	"testing"
	"time"
)

type FactSuite struct {
	EDT *time.Location
}

func Test(t *testing.T) { TestingT(t) }

var _ = Suite(&FactSuite{})

func (f *FactSuite) SetUpSuite(c *C) {
	f.EDT = time.FixedZone("EDT", -4*60*60)
}

func (f *FactSuite) TestFactFromPatient(c *C) {
	patient := LoadPatientFromFixture("../fixtures/patient-example-a.json")
	fact := FactFromPatient(patient)
	c.Assert(fact.Gender, Equals, "male")
}

func (f *FactSuite) TestFactFromMedicationStatement(c *C) {
	ms := LoadMedicationStatementFromFixture("../fixtures/medication-statement.json")
	fact := FactFromMedicationStatement(ms)
	c.Assert(fact.StartDate.Time.UnixNano(), Equals, time.Date(2015, time.April, 1, 0, 0, 0, 0, f.EDT).UnixNano())
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

func LoadMedicationStatementFromFixture(fileName string) *models.MedicationStatement {
	data, err := os.Open(fileName)
	defer data.Close()
	util.CheckErr(err)
	decoder := json.NewDecoder(data)
	ms := &models.MedicationStatement{}
	err = decoder.Decode(ms)
	util.CheckErr(err)
	return ms
}
