package storage

import "github.com/intervention-engine/ie/app"

// Interface implementation assertions
//var _ StorageService = (*mongo.MongoService)(nil)
//var _ CareTeamService = (*mongo.CareTeamService)(nil)
//var _ PatientService = (*mongo.PatientService)(nil)

type Service interface {
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
