package mongo

import (
	"github.com/intervention-engine/ie/app"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// CareTeamMembershipService mongodb membership service
type CareTeamMembershipService struct {
	C *mgo.Collection
}

// CreateCareTeamMembership create membership in mongodb
func (m *CareTeamMembershipService) CreateMembership(mem app.CareTeamMembership) error {
	return m.C.Insert(mem)
}

// PatientCareTeamMemberships list of memberships for a given care team
func (m *CareTeamMembershipService) PatientMemberships(id string) ([]app.CareTeamMembership, error) {
	var mems []app.CareTeamMembership
	err := m.C.Find(bson.M{"care_team_id": id}).All(&mems)
	return mems, err
}
