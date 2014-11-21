package models

import (
	"encoding/json"
	"github.com/pebbe/util"
	"gitlab.mitre.org/intervention-engine/fhir/models"
	. "gopkg.in/check.v1"
	"os"
	"testing"
)

type FactSuite struct {
}

func Test(t *testing.T) { TestingT(t) }

var _ = Suite(&FactSuite{})

func (s *FactSuite) TestFactFromPatient(c *C) {
	patient := LoadPatientFromFixture("../fixtures/patient-example-a.json")
	fact := FactFromPatient(patient)
	c.Assert(fact.Gender, Equals, "M")
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
