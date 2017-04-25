package mongo

import (
	"github.com/intervention-engine/ie"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// CareTeamMembershipService mongodb membership service
type CareTeamMembershipService struct {
	C *mgo.Collection
}

// CreateCareTeamMembership create membership in mongodb
func (m *CareTeamMembershipService) CreateMembership(mem ie.CareTeamMembership) error {
	return m.C.Insert(mem)
}

// PatientCareTeamMemberships list of memberships for a given care team
func (m *CareTeamMembershipService) PatientMemberships(id string) ([]ie.CareTeamMembership, error) {
	var mems []ie.CareTeamMembership
	err := m.C.Find(bson.M{"care_team_id": id}).All(&mems)
	return mems, err
}
