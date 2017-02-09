package models

import "time"

// CareTeam a collection of care providers providing care to a set of patients
type CareTeam struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string    `bson:"name,omitempty" json:"name,omitempty"`
	Leader    string    `bson:"leader,omitempty" json:"leader,omitempty"`
	CreatedAt time.Time `bson:"created_at,omitempty" json:"created_at,omitempty"`
}
