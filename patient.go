package ie

import (
	"time"

	"github.com/intervention-engine/fhir/models"
)

type PatientService interface {
	Patient(id string) (*Patient, error)
	Patients() ([]Patient, error)
	PatientsByID(ids []string) ([]Patient, error)
}

type Patient struct {
	ID                    string               `json:"id"`
	Address               Address              `json:"address"`
	Age                   int                  `json:"age"`
	Gender                string               `json:"gender"`
	BirthDate             *models.FHIRDateTime `json:"birthDate"`
	Name                  Name                 `json:"name"`
	NextHuddleID          string               `json:"nextHuddleId"`
	RecentRiskAssessments []RiskAssessment     `json:"recentRiskAssessments"`
}

type Address struct {
	Street     []string `json:"street"`
	City       string   `json:"city"`
	State      string   `json:"state"`
	PostalCode string   `json:"postalCode"`
}

type Name struct {
	Family        string `json:"family"`
	Given         string `json:"given"`
	MiddleInitial string `json:"middleInitial"`
	Full          string `json:"full"`
}

type RiskAssessment struct {
	ID      string    `json:"id"`
	GroupID string    `json:"riskAssessmentGroupId"`
	Date    time.Time `json:"date"`
	Value   int       `json:"value"`
}
