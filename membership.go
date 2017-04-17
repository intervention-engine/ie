package ie

import "time"

// Membership Represents many to many relationship between patients and care teams
type Membership struct {
	ID         string    `bson:"_id,omitempty" json:"id,omitempty"`
	CareTeamID string    `bson:"care_team_id" json:"care_team_id" binding:"required"`
	PatientID  string    `bson:"care_team_id" json:"patient_id" binding:"required"`
	CreatedAt  time.Time `bson:"created_at,omitempty" json:"created_at,omitempty"`
}

// MembershipService Manage membership in care teams for patients
type MembershipService interface {
	CreateMembership(mem Membership) error
	PatientMemberships(id string) ([]Membership, error)
}
