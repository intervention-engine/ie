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

// PatientService provides a mongo implementation of a
// Storage Service for patients.
type PatientService struct {
	S *mgo.Session
	C *mgo.Collection
}

// Patient gets a patient with the given id.
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

// Patients gets all the patients in the db.
func (s *PatientService) Patients() ([]*app.Patient, error) {
	defer s.S.Close()
	var data []Patient
	err := s.C.Find(nil).All(&data)
	if err != nil {
		return nil, err
	}

	pp := make([]*app.Patient, len(data), len(data))
	for i, _ := range data {
		pp[i] = newPatient(data[i])
	}

	return pp, nil
}

// SortBy gets patients sorted by the fields given.
func (s *PatientService) SortBy(fields ...string) ([]*app.Patient, error) {
	defer s.S.Close()
	var data []Patient
	err := s.C.Find(nil).Sort(fields...).All(&data)
	if err != nil {
		return nil, err
	}

	pp := make([]*app.Patient, len(data))
	for i, patient := range data {
		pp[i] = newPatient(patient)
	}

	return pp, nil
}

// Patient embeds FHIR model and adds Risk/Huddle information
type Patient struct {
	models.Patient  `bson:",inline"`
	RiskAssessments []app.RiskAssessment `bson:"risk_assessment,omitempty" json:"risk_assessment,omitempty"`
	NextHuddleID    string               `bson:"next_huddle_id,omitempty" json:"next_huddle_id,omitempty"`
}

func newPatient(fhirPatient Patient) *app.Patient {
	p := app.Patient{}
	p.ID = fhirPatient.Id
	if len(fhirPatient.Address) > 0 {
		p.Address = newAddress(fhirPatient.Address[0])
	}
	if fhirPatient.BirthDate != nil {
		age := age(fhirPatient.BirthDate)
		p.Age = &age
	}
	p.Gender = &fhirPatient.Gender
	p.BirthDate = &fhirPatient.BirthDate.Time
	if len(fhirPatient.Name) > 0 {
		p.Name = newName(fhirPatient.Name[0])
	}
	// p.NextHuddleID = &fhirPatient.NextHuddleID
	// p.RecentRiskAssessments = fhirPatient.RiskAssessments
	return &p
}

func newName(name models.HumanName) *app.Name {
	n := app.Name{}
	var family, given string
	if len(name.Given) > 0 {
		given = name.Given[0]
		n.Given = &given
	}
	if len(name.Family) > 0 {
		family = name.Family[0]
		n.Family = &family
	}
	if (given != "") && (family != "") {
		full := fmt.Sprintf("%s %s", given, family)
		n.Full = &full
	} else if given != "" {
		n.Full = &given
	} else if family != "" {
		n.Full = &family
	}
	return &n
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
