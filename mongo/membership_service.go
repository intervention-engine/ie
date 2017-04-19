package mongo

import (
	"github.com/intervention-engine/ie"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// MembershipService mongodb membership service
type MembershipService struct {
	C *mgo.Collection
}

// CreateMembership create membership in mongodb
func (m *MembershipService) CreateMembership(mem ie.Membership) error {
	return m.C.Insert(mem)
}

// PatientMemberships list of memberships for a given care team
func (m *MembershipService) PatientMemberships(id string) ([]ie.Membership, error) {
	var mems []ie.Membership
	err := m.C.Find(bson.M{"care_team_id": id}).All(&mems)
	return mems, err
}
