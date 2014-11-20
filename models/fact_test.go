package models

import (
	"gitlab.mitre.org/intervention-engine/fhir/models"
	. "gopkg.in/check.v1"
	"testing"
)

type FactSuite struct {
}

func Test(t *testing.T) { TestingT(t) }

var _ = Suite(&FactSuite{})

func (s *FactSuite) TestFactFromPatient(c *C) {
	patient := &models.Patient{}
	patient.Gender = models.CodeableConcept{Coding: []models.Coding{models.Coding{System: "http://hl7.org/fhir/v3/AdministrativeGender", Code: "M"}}}
	fact := FactFromPatient(patient)
	c.Assert(fact.Gender, Equals, "M")
}
