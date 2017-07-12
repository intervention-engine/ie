package storage

import (
	"time"

	"github.com/intervention-engine/ie/app"
)

// Service is a description of a factory that the app expects a storage package to
// implement in order for the application to be able to access its resources particular
// service.
type ServiceFactory interface {
	NewCareTeamService() CareTeamService
	NewPatientService() PatientService
	NewHuddleService() HuddleService
	NewRiskAssessmentService() RiskAssessmentService
}

// CareTeamService describes the interface for storing a CareTeam
type CareTeamService interface {
	CareTeam(id string) (*app.CareTeam, error)
	CareTeams() ([]*app.CareTeam, error)
	CreateCareTeam(c *app.CareTeam) error
	UpdateCareTeam(c *app.CareTeam) error
	DeleteCareTeam(id string) error
	AddPatient(careTeamID string, patientID string) error
	RemovePatient(careTeamID string, patientID string) error
}

// RiskAssessmentService List Risk assessments for a patient and risk service. Optionally query on a time period
type RiskAssessmentService interface {
	RiskAssessment(id string) (*app.RiskAssessment, error)
	RiskAssessments(patientID string, serviceID string, start time.Time, end time.Time) ([]*app.RiskAssessment, error)
}

// PatientService describes the interface for storing a Patient
type PatientService interface {
	Patient(id string) (*app.Patient, error)
	Patients() ([]*app.Patient, error)
	PatientsSortBy(...string) ([]*app.Patient, error)
}

// HuddleService describes the interface for storing a huddle.
type HuddleService interface {
	HuddlesFilterBy(query HuddleFilterQuery) ([]*app.Huddle, error)
	ScheduleHuddle(careTeamID string, patientID string, date time.Time) (*app.Huddle, bool, error)
	DeletePatient(huddleID string, patientID string) (*app.Huddle, error)
}

var AcceptedQueryFields = []string{
	"name",
	"-name",
	"birthdate",
	"-birthdate",
	"gender",
	"-gender",
	//"address",
	//"-address",
	//"riskScore",
	//"-riskScore",
	//"notifications",
	//"-notifications",
}

type HuddleFilterQuery struct {
	CareTeamID string
	PatientID  string
	Date       time.Time
}
