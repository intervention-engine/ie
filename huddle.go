package ie

import (
	"time"
)

const (
	RISK_SCORE = iota
	EVENT      = iota
	MANUAL     = iota
	ROLLOVER   = iota
)

type HuddleService interface {
	CareTeamHuddles(id string) ([]Huddle, error)
	UpsertHuddles(huddles []Huddle) error
}

type Huddle struct {
	ID         string
	Date       time.Time
	CareTeamID string
}

type HuddleMembership struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	HuddleID  string    `bson:"huddle_id" json:"huddle_id" binding:"required"`
	PatientID string    `bson:"care_team_id" json:"patient_id" binding:"required"`
	CreatedAt time.Time `bson:"created_at,omitempty" json:"created_at,omitempty"`
	Reason    string    `bson:"reason,omitempty" json:"reason,omitempty"`
	Reviewed  bool      `bson:"reviewed,omitempty" json:"reason,omitempty"`
}

// CareTeamMembershipService Manage membership in care teams for patients
type HuddleMembershipService interface {
	CreateMembership(mem HuddleMembership) error
	DeleteMembership(id string) error
	HuddleMemberships(id string) ([]HuddleMembership, error)
	PatientMemberships(id string) ([]HuddleMembership, error)
}
