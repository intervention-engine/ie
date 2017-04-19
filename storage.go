package ie

import "github.com/intervention-engine/ie/app"

type StorageService interface {
	NewCareTeamService() CareTeamService
	NewPatientService() PatientService
}

// CareTeamService describes the interface for storing a CareTeam
type CareTeamService interface {
	CareTeam(id string) (*app.Careteam, error)
	CareTeams() ([]*app.Careteam, error)
	CreateCareTeam(c *app.Careteam) error
	UpdateCareTeam(c *app.Careteam) error
	DeleteCareTeam(id string) error
}

type PatientService interface {
	Patient(id string) (*app.Patient, error)
	Patients() ([]*app.Patient, error)
}
