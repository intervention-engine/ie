package models

import "github.com/intervention-engine/fhir/models"

// Patient embeds FHIR model and adds Risk/Huddle information
type Patient struct {
	models.Patient  `bson:",inline"`
	RiskAssessments []RiskAssessment `bson:"risk_assessment,omitempty" json:"risk_assessment,omitempty"`
	NextHuddleID    string           `bson:"next_huddle_id,omitempty" json:"next_huddle_id,omitempty"`
}

// // NextHuddle Get the Next Huddle this patient is in
// func (p *Patient) NextHuddle() *Huddle {
//
// }
