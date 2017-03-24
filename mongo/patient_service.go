package mongo

import (
	"errors"
	"fmt"
	"time"

	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/ie"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type PatientService struct {
	C *mgo.Collection
}

// Patient embeds FHIR model and adds Risk/Huddle information
type Patient struct {
	models.Patient  `bson:",inline"`
	RiskAssessments []ie.RiskAssessment `bson:"risk_assessment,omitempty" json:"risk_assessment,omitempty"`
	NextHuddleID    string              `bson:"next_huddle_id,omitempty" json:"next_huddle_id,omitempty"`
}

func (s *PatientService) Patient(id string) (*ie.Patient, error) {
	if !bson.IsObjectIdHex(id) {
		return nil, errors.New("bad id")
	}
	var data Patient
	err := s.C.FindId(id).One(&data)
	if err != nil {
		return nil, err
	}

	p := newPatient(data)

	return &p, nil
}

func (s *PatientService) Patients() ([]ie.Patient, error) {
	var data []Patient
	err := s.C.Find(nil).All(&data)
	if err != nil {
		return nil, err
	}

	pp := make([]ie.Patient, len(data))
	for i, patient := range data {
		pp[i] = newPatient(patient)
	}

	return pp, nil
}

func newPatient(fhirPatient Patient) ie.Patient {
	p := ie.Patient{}
	p.ID = fhirPatient.Id
	p.Address = newAddress(fhirPatient.Address[0])
	p.Age = age(fhirPatient.BirthDate)
	p.Gender = fhirPatient.Gender
	p.BirthDate = fhirPatient.BirthDate
	p.Name = newName(fhirPatient.Name[0])
	p.NextHuddleID = fhirPatient.NextHuddleID
	p.RecentRiskAssessments = fhirPatient.RiskAssessments
	return p
}

func newAddress(address models.Address) ie.Address {
	a := ie.Address{}
	a.Street = address.Line
	a.City = address.City
	a.State = address.State
	a.PostalCode = address.PostalCode
	return a
}

func newName(name models.HumanName) ie.Name {
	n := ie.Name{}
	n.Family = name.Family[0]
	n.Given = name.Given[0]
	n.Full = fmt.Sprintf("%s %s", n.Given, n.Family)
	return n
}

func age(birthday *models.FHIRDateTime) int {
	now := time.Now()
	years := now.Year() - birthday.Time.Year()

	if now.YearDay() < birthday.Time.YearDay() {
		years--
	}

	return years
}
