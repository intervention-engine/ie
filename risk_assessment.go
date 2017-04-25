package ie

import "time"

type RiskService struct {
	ID                 string
	Name               string
	URL                string
	RiskCategoryGroups []RiskCategoryGroup
}

type RiskCategoryGroup struct {
	ID       string
	Name     string
	Weight   float32
	MaxValue float32
}

type RiskAssessment struct {
	ID             string    `json:"id"`
	ServiceID      string    `json:"service_id"`
	Date           time.Time `json:"date"`
	Value          float32   `json:"value"`
	RiskCategories []RiskCategory
}

type RiskCategory struct {
	ID                  string
	RiskCategoryGroupID string
	Value               float32
}

// GET /api/risk_services
// GET /api/patient/:patient_id/risk_assessments/:id
// GET /api/risk_assessment/:id/breakdown
