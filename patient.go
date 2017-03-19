package ie

import "github.com/intervention-engine/fhir/models"

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
	ID      string            `json:"id"`
	Address RestructedAddress `json:"address"`
}

func (p *RestructedPatient) FromFHIR(patient *models.Patient) *RestructedPatient {
	p.ID = patient.Id
	p.Address = *(&RestructedAddress{}).FromFHIR(&patient.Address[0])
	return p
}
