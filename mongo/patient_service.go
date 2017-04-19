package mongo

import (
	"errors"
	"fmt"
	"time"

	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/ie/app"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type PatientService struct {
	S *mgo.Session
	C *mgo.Collection
}

// Patient embeds FHIR model and adds Risk/Huddle information
type Patient struct {
	models.Patient  `bson:",inline"`
	RiskAssessments []app.RiskAssessment `bson:"risk_assessment,omitempty" json:"risk_assessment,omitempty"`
	NextHuddleID    string               `bson:"next_huddle_id,omitempty" json:"next_huddle_id,omitempty"`
}

func (s *PatientService) Patient(id string) (*app.Patient, error) {
	defer s.S.Close()
	if !bson.IsObjectIdHex(id) {
		return nil, errors.New("bad id")
	}
	var data Patient
	err := s.C.FindId(id).One(&data)
	if err != nil {
		return nil, err
	}

	p := newPatient(data)

	return p, nil
}

func (s *PatientService) Patients() ([]*app.Patient, error) {
	defer s.S.Close()
	var data []Patient
	err := s.C.Find(nil).All(&data)
	if err != nil {
		return nil, err
	}

	pp := make([]*app.Patient, len(data))
	for i, patient := range data {
		pp[i] = newPatient(patient)
	}

	return pp, nil
}

func newPatient(fhirPatient Patient) *app.Patient {
	p := app.Patient{}
	p.ID = fhirPatient.Id
	p.Address = newAddress(fhirPatient.Address[0])
	age := age(fhirPatient.BirthDate)
	p.Age = &age
	p.Gender = &fhirPatient.Gender
	// TODO: do this conversion
	// p.BirthDate = &fhirPatient.BirthDate
	name := fhirPatient.Name[0]
	p.Name.Family = &name.Family[0]
	p.Name.Given = &name.Given[0]
	full := fmt.Sprintf("%s %s", name.Given, name.Family)
	p.Name.Full = &full
	// p.NextHuddleID = &fhirPatient.NextHuddleID
	// p.RecentRiskAssessments = fhirPatient.RiskAssessments
	return &p
}

func newAddress(address models.Address) *app.Address {
	a := app.Address{}
	a.Street = address.Line
	a.City = &address.City
	a.State = &address.State
	a.PostalCode = &address.PostalCode
	return &a
}

func age(birthday *models.FHIRDateTime) int {
	now := time.Now()
	years := now.Year() - birthday.Time.Year()

	if now.YearDay() < birthday.Time.YearDay() {
		years--
	}

	return years
}
