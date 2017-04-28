package storage

import "github.com/intervention-engine/ie/app"

// Service is a description of a factory that the app expects a storage package to
// implement in order for the application to be able to access its resources particular
// service.
type Service interface {
	NewCareTeamService() CareTeamService
	NewPatientService() PatientService
}

// CareTeamService describes the interface for storing a CareTeam
type CareTeamService interface {
	CareTeam(id string) (*app.CareTeam, error)
	CareTeams() ([]*app.CareTeam, error)
	CreateCareTeam(c *app.CareTeam) error
	UpdateCareTeam(c *app.CareTeam) error
	DeleteCareTeam(id string) error
}

type HuddleService interface {
	CareTeamHuddles(id string) ([]app.Huddle, error)
	UpsertHuddles(huddles []app.Huddle) error
}

// PatientService describes the interface for storing a Patient
type PatientService interface {
	Patient(id string) (*app.Patient, error)
	Patients() ([]*app.Patient, error)
	SortBy(...string) ([]*app.Patient, error)
}

type HuddleMembershipService interface {
	CreateMembership(mem app.HuddleMembership) error
	DeleteMembership(id string) error
	HuddleMemberships(id string) ([]app.HuddleMembership, error)
	PatientMemberships(id string) ([]app.HuddleMembership, error)
}

// CareTeamMembershipService Manage membership in care teams for patients
type CareTeamMembershipService interface {
	CreateMembership(mem app.CareTeamMembership) error
	PatientMemberships(id string) ([]app.CareTeamMembership, error)
}
