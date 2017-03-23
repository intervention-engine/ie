package ie

import "time"

// RiskAssessment represents a single score for a risk algorithm
type RiskAssessment struct {
	ID      string    `json:"id,omitempty" bson:"_id,omitempty"`
	GroupID string    `json:"risk_assessment_group_id,omitempty"`
	Date    time.Time `json:"date,omitempty"`
	Value   int       `json:"value,omitempty"`
}

type RestructedRiskAssessment struct {
	ID                    string `json:"id"`
	RiskAssessmentGroupId string `json:"riskAssessmentGroupId"`
	Date                  string `json:"date"`
	Value                 int    `json:"value"`
}
