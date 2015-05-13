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

type FactSuite struct{}

func Test(t *testing.T) { TestingT(t) }

var _ = Suite(&FactSuite{})

func (f *FactSuite) TestFactFromPatient(c *C) {
	patient := LoadPatientFromFixture("../fixtures/patient-example-a.json")
	fact := FactFromPatient(patient)
	c.Assert(fact.Gender, Equals, "M")
}

func (f *FactSuite) TestFactFromMedicationStatement(c *C) {
	ms := LoadMedicationStatementFromFixture("../fixtures/medication-statement.json")
	mlu := func(id string) (models.Medication, error) {
		c.Assert(id, Equals, "5540f2041cd462313300000c")
		coding := models.Coding{System: "Foo", Code: "Bar"}
		return models.Medication{Code: &models.CodeableConcept{Coding: []models.Coding{coding}}}, nil
	}
	fact := FactFromMedicationStatement(ms, mlu)
	c.Assert(fact.StartDate.Time, Equals, time.Date(2014, time.January, 1, 0, 0, 0, 0, time.UTC))
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
