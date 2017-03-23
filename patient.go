package ie

import (
	"fmt"
	"time"

	"github.com/intervention-engine/fhir/models"
)

// Patient embeds FHIR model and adds Risk/Huddle information
type Patient struct {
	models.Patient  `bson:",inline"`
	RiskAssessments []RiskAssessment `bson:"risk_assessment,omitempty" json:"risk_assessment,omitempty"`
	NextHuddleID    string           `bson:"next_huddle_id,omitempty" json:"next_huddle_id,omitempty"`
}

type PatientService interface {
	Patient(id string) (*Patient, error)
	Patients() ([]RestructedPatient, error)
}

type RestructedPatient struct {
	ID                    string                     `json:"id"`
	Address               RestructedAddress          `json:"address"`
	Age                   int                        `json:"age"`
	Gender                string                     `json:"gender"`
	BirthDate             *models.FHIRDateTime       `json:"birthDate"`
	Name                  RestructedPatientName      `json:"name"`
	NextHuddleID          *string                    `json:"nextHuddleId"`
	RecentRiskAssessments []RestructedRiskAssessment `json:"recentRiskAssessments"`
}

func (p *RestructedPatient) FromFHIR(patient *models.Patient) *RestructedPatient {
	p.ID = patient.Id
	p.Address = *(&RestructedAddress{}).FromFHIR(&patient.Address[0])
	p.Age = age(patient.BirthDate)
	p.Gender = patient.Gender
	p.BirthDate = patient.BirthDate
	p.Name = *(&RestructedPatientName{}).FromFHIR(&patient.Name[0])
	return p
}

type RestructedPatientName struct {
	Family        string `json:"family"`
	Given         string `json:"given"`
	MiddleInitial string `json:"middleInitial"`
	Full          string `json:"full"`
}

func (p *RestructedPatientName) FromFHIR(name *models.HumanName) *RestructedPatientName {
	p.Family = name.Family[0]
	p.Given = name.Given[0]
	p.Full = fmt.Sprintf("%s %s", p.Given, p.Family)
	return p
}

func age(birthday *models.FHIRDateTime) int {
	now := time.Now()
	years := now.Year() - birthday.Time.Year()

	if now.YearDay() < birthday.Time.YearDay() {
		years--
	}

	return years
}
