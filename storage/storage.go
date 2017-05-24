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

// PatientService describes the interface for storing a Patient
type PatientService interface {
	Patient(id string) (*app.Patient, error)
	Patients() ([]*app.Patient, error)
	PatientsSortBy(...string) ([]*app.Patient, error)
}

// HuddleService describes the interface for storing a huddle.
type HuddleService interface {
	Huddles() ([]*app.Huddle, error)
	HuddlesForCareTeam(id string) ([]*app.Huddle, error)
	ScheduleHuddle(careTeamID string, patientID string, date time.Time) (*app.Huddle, bool, error)
	DeletePatient(huddleID string, patientID string) (*app.Huddle, error)
}

var AcceptedQueryFields = []string{
	//"+name",
	"name",
	"-name",
	//"+birthdate",
	"birthdate",
	"-birthdate",
	//"+gender",
	"gender",
	"-gender",
	//"+address",
	"address",
	"-address",
	//"+riskScore",
	//"riskScore",
	//"-riskScore",
	//"+notifications",
	//"notifications",
	//"-notifications",
}
