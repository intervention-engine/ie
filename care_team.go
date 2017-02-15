package ie

import "time"

// CareTeam a collection of care providers providing care to a set of patients
type CareTeam struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string    `form:"name" bson:"name,omitempty" json:"name,omitempty" binding:"required"`
	Leader    string    `form:"leader" bson:"leader,omitempty" json:"leader,omitempty" binding:"required"`
	CreatedAt time.Time `bson:"created_at,omitempty" json:"created_at,omitempty"`
}

// CareTeamService describes the interface for storing a CareTeam
type CareTeamService interface {
	CareTeam(id string) (*CareTeam, error)
	CareTeams() ([]CareTeam, error)
	CreateCareTeam(c *CareTeam) error
	UpdateCareTeam(c *CareTeam) error
	DeleteCareTeam(id string) error
}
